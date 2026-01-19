package main

import (
	"bytes"
	"encoding/gob"
	"file_distribution_system/p2p"
	"fmt"
	"io"
	"log"
	"sync"
	"time"
)

type FileServerOpts struct {
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
}

func NewFile(opts FileServerOpts) *FileServer {
	storeOpts := StoreOpts{
		Root:              opts.StorageRoot,
		PathTransformFunc: opts.PathTransformFunc,
	}
	return &FileServer{
		FileServerOpts: opts,
		store:          NewStore(storeOpts),
		quitchan:       make(chan struct{}),
		peers:          make(map[string]p2p.Peer),
	}
}

func (s *FileServer) broadcast(msg *Message) error {
	peers := []io.Writer{}
	for _, peer := range s.peers {
		peers = append(peers, peer)
	}
	mw := io.MultiWriter(peers...)
	return gob.NewEncoder(mw).Encode(msg)
}

type Message struct {
	Payload any
}

type MessageStoreFile struct {
	Key  string
	Size int64
}
//Store this file to disc
	// broadcast this file to all known peers in the network
func (s *FileServer) StoreData(key string, r io.Reader) error {
	
	buf := new(bytes.Buffer)
	tee := io.TeeReader(r, buf)
	size, err := s.store.Write(key, tee)
	if err != nil {
		return err
	}

	msg := Message{
		Payload: MessageStoreFile{
			Key:  key,
			Size: size,
		},
	}
	msgBuf := new(bytes.Buffer)

	if err := gob.NewEncoder(msgBuf).Encode(msg); err != nil {
		return err
	}

	for _, peer := range s.peers {
		if err := peer.Send(msgBuf.Bytes()); err != nil {
			return err
		}
	}

	time.Sleep(time.Second * 3)

	for _, peer := range s.peers {
		n, err := io.Copy(peer, buf)
		if err != nil {
			return err
		}
		fmt.Println("recieved and written bytes to disk: ", n)
	}

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

func (s *FileServer) loop() {
	defer func() {
		log.Println("file server stopped due to user quit action")
		s.Transport.Close()
	}()
	for {
		select {
		case rpc := <-s.Transport.Consume():
			var msg Message
			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&msg); err != nil {
				log.Println(err)
				return
			}
			if err := s.handleMessage(rpc.From, &msg); err != nil {
				log.Println(err)
				return
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
	}

	return nil
}

func (s *FileServer) handleMessageStoreFile(from string, msg MessageStoreFile) error {
	peer, ok := s.peers[from]
	if !ok {
		return fmt.Errorf("peer (%s) could not be found in the peer list", from)
	}
	if _, err := s.store.Write(msg.Key, io.LimitReader(peer, msg.Size)); err != nil {
		return err
	}
	peer.(*p2p.TCPPeer).Wg.Done()
	return nil
}

func (s *FileServer) bootstrapNetwrk() error {
	for _, addr := range s.BootstrapNodes {
		if len(addr) == 0 {
			continue
		}
		go func(addr string) {
			fmt.Println(" attempting to connect with remote:", addr)
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
}
