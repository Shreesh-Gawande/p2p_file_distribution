package main

import (
	"file_distribution_system/p2p"
	"fmt"
	"log"
	"sync"
)

type FileServerOpts struct {
	StorageRoot string
	PathTransformFunc PathTransformFunc
	Transport         p2p.Transport
	BootstrapNodes    [] string
}

type FileServer struct {
	FileServerOpts
     
	peerlock   sync.Mutex
	peers      map[string]p2p.Peer

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

func (s *FileServer) Stop() {
	close(s.quitchan)
}

func(s *FileServer) OnPeer(p p2p.Peer) error{
	s.peerlock.Lock()
	defer s.peerlock.Unlock()

	s.peers[p.RemoteAddr().String()]=p

	log.Printf(" connected with remote %s", p.RemoteAddr())
	return nil
}

func (s *FileServer) loop() {

	defer func(){
      log.Println("file server stopped due to user quit action")
	  s.Transport.Close()
	}()
	for {
		select {
		case msg := <-s.Transport.Consume():
			fmt.Println(msg)
		case <-s.quitchan:
			return
		}
	}
}
func(s *FileServer) bootstrapNetwrk() error{
	for _, addr :=range s.BootstrapNodes{
		if len(addr)==0{
			continue
		}
		go func(addr string){
			fmt.Println(" attempting to connect with remote:", addr)
			if err :=s.Transport.Dial(addr); err!=nil{
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
