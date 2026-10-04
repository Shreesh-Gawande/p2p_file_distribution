package main

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"file_distribution_system/p2p"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type FileServerOpts struct {
	ID                string
	EncKey            []byte
	StorageRoot       string
	PathTransformFunc PathTransformFunc
	Transport         p2p.Transport
	BootstrapNodes    []string
}

type FileServer struct {
	FileServerOpts

	peerlock sync.Mutex
	peers    map[string]p2p.Peer

	store    *Storage
	quitchan chan struct{}
	acks     chan struct{}
}

func NewFile(opts FileServerOpts) *FileServer {
	storeOpts := StoreOpts{
		Root:              opts.StorageRoot,
		PathTransformFunc: opts.PathTransformFunc,
	}
	if len(opts.ID) == 0 {
		opts.ID = generateID()
	}
	return &FileServer{
		FileServerOpts: opts,
		store:          NewStore(storeOpts),
		quitchan:       make(chan struct{}),
		acks:           make(chan struct{}, 32),
		peers:          make(map[string]p2p.Peer),
	}
}

func encodeMessage(msg *Message) ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		return nil, err
	}
	frame := new(bytes.Buffer)
	frame.WriteByte(p2p.IncomingMessage)
	if err := binary.Write(frame, binary.LittleEndian, uint32(buf.Len())); err != nil {
		return nil, err
	}
	if _, err := frame.Write(buf.Bytes()); err != nil {
		return nil, err
	}
	return frame.Bytes(), nil
}

func sendMessage(peer p2p.Peer, msg *Message) error {
	frame, err := encodeMessage(msg)
	if err != nil {
		return err
	}
	return peer.Send(frame)
}

func (s *FileServer) broadcast(msg *Message) error {
	s.peerlock.Lock()
	peers := make([]p2p.Peer, 0, len(s.peers))
	for _, peer := range s.peers {
		peers = append(peers, peer)
	}
	s.peerlock.Unlock()
	for _, peer := range peers {
		if err := sendMessage(peer, msg); err != nil {
			return err
		}
	}

	return nil
}

type Message struct {
	Payload any
}

type MessageStoreFile struct {
	ID       string
	Key      string
	FileName string
	Size     int64
}

type MessageGetFile struct {
	ID  string
	Key string
}

type MessageFileReceived struct{}

func (s *FileServer) Get(key string) (io.Reader, error) {
	if s.store.Has(s.ID, key) {
		fmt.Printf("[%s]serving file(%s) from local disk\n", s.Transport.Addr(), key)
		_, r, err := s.store.Read(s.ID, key)
		return r, err
	}

	fmt.Printf("[%s]dont have file (%s) locally, fetching from network ...\n", s.Transport.Addr(), key)
	msg := Message{
		Payload: MessageGetFile{
			Key: hashKey(key),
			ID:  s.ID,
		},
	}
	if err := s.broadcast(&msg); err != nil {
		return nil, err
	}
	time.Sleep(time.Millisecond * 500)
	for _, peer := range s.peers {
		peer.WaitForStream()

		//First read the file size so we can limit the amount of bytes that we read from the connection
		// //, so it will not keep hanging
		var fileSize int64
		binary.Read(peer, binary.LittleEndian, &fileSize)
		n, err := s.store.WriteDecrypt(s.ID, s.EncKey, key, io.LimitReader(peer, fileSize))

		if err != nil {
			return nil, err
		}
		fmt.Printf("[%s] recieved(%d) bytes over the network from (%s) \n ", s.Transport.Addr(), n, peer.RemoteAddr())

		peer.CloseStream()
	}

	_, r, err := s.store.Read(s.ID, key)
	return r, err
}

// Store this file to disc
// broadcast this file to all known peers in the network
func (s *FileServer) Store(key string, r io.Reader) error {

	s.peerlock.Lock()
	peers := make([]p2p.Peer, 0, len(s.peers))
	for _, peer := range s.peers {
		peers = append(peers, peer)
	}
	s.peerlock.Unlock()
	if len(peers) == 0 {
		return fmt.Errorf("no connected peers to send the file to")
	}

	fileBuffer := new(bytes.Buffer)
	tee := io.TeeReader(r, fileBuffer)
	size, err := s.store.Write(s.ID, key, tee)
	if err != nil {
		return err
	}

	msg := Message{
		Payload: MessageStoreFile{
			ID:       s.ID,
			Key:      hashKey(key),
			FileName: filepath.Base(key),
			Size:     size,
		},
	}

	if err := s.broadcast(&msg); err != nil {
		return err
	}

	for _, peer := range peers {
		if err := peer.Send([]byte{p2p.IncomingStream}); err != nil {
			return err
		}
		if _, err := io.Copy(peer, bytes.NewReader(fileBuffer.Bytes())); err != nil {
			return err
		}
	}
	for range peers {
		select {
		case <-s.acks:
		case <-time.After(5 * time.Minute):
			return fmt.Errorf("timed out waiting for the receiving peer to confirm the file")
		}
	}

	fmt.Printf("[%s]sent %d bytes to %d peer(s)\n", s.Transport.Addr(), size, len(peers))

	return nil

}

func (s *FileServer) Stop() {
	close(s.quitchan)
}

func (s *FileServer) OnPeer(p p2p.Peer) error {
	s.peerlock.Lock()
	defer s.peerlock.Unlock()

	s.peers[p.RemoteAddr().String()] = p
	log.Printf("✓ Peer connected: %s -> %s\n", s.StorageRoot, p.RemoteAddr())
	return nil
}

func (s *FileServer) WaitForPeer(timeout time.Duration) error {
	deadline := time.NewTimer(timeout)
	defer deadline.Stop()
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		s.peerlock.Lock()
		hasPeer := len(s.peers) > 0
		s.peerlock.Unlock()
		if hasPeer {
			return nil
		}

		select {
		case <-deadline.C:
			return fmt.Errorf("timed out waiting for a peer connection")
		case <-ticker.C:
		}
	}
}

func (s *FileServer) loop() {
	defer func() {
		log.Println("file server stopped due to error or user quit action")
		s.Transport.Close()
	}()
	for {
		select {
		case rpc := <-s.Transport.Consume():
			var msg Message
			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&msg); err != nil {
				log.Println("decoding error: ", err)
				continue
			}
			if err := s.handleMessage(rpc.From, &msg); err != nil {
				log.Println("handle message error : ", err)
			}
		case <-s.quitchan:
			return
		}
	}
}

func (s *FileServer) handleMessage(from string, msg *Message) error {
	switch v := msg.Payload.(type) {
	case MessageStoreFile:
		return s.handleMessageStoreFile(from, v)
	case MessageGetFile:
		return s.handelMessageGetFile(from, v)
	case MessageFileReceived:
		s.acks <- struct{}{}
	}

	return nil
}

func (s *FileServer) handelMessageGetFile(from string, msg MessageGetFile) error {
	if !s.store.Has(msg.ID, msg.Key) {
		return fmt.Errorf("[%s] need to serve file (%s) but it did not exist on disk", s.Transport.Addr(), msg.Key)
	}
	fmt.Printf("[%s]serving file (%s) over the network \n", s.Transport.Addr(), msg.Key)

	fileSize, r, err := s.store.Read(msg.ID, msg.Key)
	if err != nil {
		return err
	}

	if rc, ok := r.(io.ReadCloser); ok {
		fmt.Println("closing ReadCloser")
		defer rc.Close()
	}

	peer, ok := s.peers[from]

	if !ok {
		return fmt.Errorf("peer %s not in map", from)
	}
	peer.WaitForStream()
	//First send the "incommingStream" byte to the peer and then we can send the file size as an int64
	peer.Send([]byte{p2p.IncomingStream})

	binary.Write(peer, binary.LittleEndian, fileSize)

	n, err := io.Copy(peer, r)
	if err != nil {
		return err
	}

	fmt.Printf("[%s]written %d bytes over the network to %s \n", s.Transport.Addr(), n, from)

	return nil

}

func (s *FileServer) handleMessageStoreFile(from string, msg MessageStoreFile) error {
	peer, ok := s.peers[from]
	if !ok {
		return fmt.Errorf("peer (%s) could not be found in the peer list", from)
	}
	fileName := filepath.Base(msg.FileName)
	if fileName == "." || fileName == string(filepath.Separator) || fileName == "" {
		return fmt.Errorf("invalid incoming file name %q", msg.FileName)
	}
	peer.WaitForStream()
	if err := os.MkdirAll("received_files", 0o755); err != nil {
		return err
	}
	file, err := os.Create(filepath.Join("received_files", fileName))
	if err != nil {
		return err
	}
	n, copyErr := io.CopyN(file, peer, msg.Size)
	closeErr := file.Close()
	if copyErr != nil {
		return copyErr
	}
	if closeErr != nil {
		return closeErr
	}
	fmt.Printf("[%s]written %d bytes on disc:\n", s.Transport.Addr(), n)
	peer.CloseStream()
	return sendMessage(peer, &Message{Payload: MessageFileReceived{}})
}

func (s *FileServer) bootstrapNetwrk() error {
	for _, addr := range s.BootstrapNodes {
		if len(addr) == 0 {
			continue
		}
		go func(addr string) {
			fmt.Printf("[%s] attempting to connect with remote: %s\n", s.Transport.Addr(), addr)
			if err := s.Transport.Dial(addr); err != nil {
				log.Println(" dial error :", err)
			}
		}(addr)
	}

	return nil
}

func (s *FileServer) Start() error {
	if err := s.Transport.ListenAndAccept(); err != nil {
		return err
	}
	s.bootstrapNetwrk()
	s.loop()
	return nil
}

func init() {
	gob.Register(MessageStoreFile{})
	gob.Register(MessageGetFile{})
	gob.Register(MessageFileReceived{})
}
