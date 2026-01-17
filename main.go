package main

import (
	"file_distribution_system/p2p"
	"log"
)

func makeserver(listenAddr string, nodes ...string) *FileServer{
		tcpTransportOpts:=p2p.TCPTransportOpts{
		ListenAddr: listenAddr,
		Handshake:  p2p.NOPHandshakeFunc,
		Decoder:    p2p.DefaultDecoder{},
		//TODO onPeer func
	}

	tcpTransport:=p2p.NewTCPTransport(tcpTransportOpts)

	fileTransport:=FileServerOpts{
		StorageRoot:       listenAddr+"_network",
		PathTransformFunc: CASPathTransformFunc,
		Transport:      tcpTransport,   
		BootstrapNodes: nodes,
	}
	s:= NewFile(fileTransport)
    tcpTransport.OnPeer=s.OnPeer
	return s
}

func main() {
    s1:=makeserver(":3000", "")
    s2:=makeserver(":4000", ":3000")
	go func() {
     log.Fatal(s1.Start())
	}()

	s2.Start()
}
