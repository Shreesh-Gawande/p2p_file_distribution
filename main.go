package main

import (
	"file_distribution_system/p2p"
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func makeserver(listenAddr string, nodes ...string) *FileServer {
	tcpTransportOpts := p2p.TCPTransportOpts{
		ListenAddr: listenAddr,
		Handshake:  p2p.NOPHandshakeFunc,
		Decoder:    p2p.DefaultDecoder{},
		//TODO onPeer func
	}

	tcpTransport := p2p.NewTCPTransport(tcpTransportOpts)
	folderName := strings.NewReplacer(":", "_", ".", "_").Replace(strings.TrimPrefix(listenAddr, ":"))
	fileTransport := FileServerOpts{
		EncKey:            newEncryptionKey(),
		StorageRoot:       folderName + "_network",
		PathTransformFunc: CASPathTransformFunc,
		Transport:         tcpTransport,
		BootstrapNodes:    nodes,
	}
	s := NewFile(fileTransport)
	tcpTransport.OnPeer = s.OnPeer
	return s
}

func main() {
	listenAddr := flag.String("listen", ":3000", "TCP address to listen on (for example :3000)")
	peerAddr := flag.String("peer", "", "optional peer address to connect to (for example 192.168.1.20:3000)")
	sendPath := flag.String("send", "", "optional path of a file to send to connected peers")
	flag.Parse()

	server := makeserver(*listenAddr, *peerAddr)
	log.Printf("Starting file-sharing node on %s", *listenAddr)
	startErr := make(chan error, 1)
	go func() {
		startErr <- server.Start()
	}()

	if *sendPath == "" {
		if err := <-startErr; err != nil {
			log.Fatal(err)
		}
		return
	}

	if err := server.WaitForPeer(30 * time.Second); err != nil {
		server.Stop()
		<-startErr
		log.Fatal(err)
	}

	file, err := os.Open(*sendPath)
	if err != nil {
		server.Stop()
		<-startErr
		log.Fatal(err)
	}

	if err := server.Store(filepath.Base(*sendPath), file); err != nil {
		server.Stop()
		<-startErr
		log.Fatal(err)
	}
	if err := file.Close(); err != nil {
		server.Stop()
		<-startErr
		log.Fatal(err)
	}
	log.Printf("Sent %s to connected peer(s)", filepath.Base(*sendPath))
	server.Stop()
	if err := <-startErr; err != nil {
		log.Fatal(err)
	}
}
