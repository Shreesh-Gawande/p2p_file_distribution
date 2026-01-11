package main

import (
	"file_distribution_system/p2p"
	"fmt"

	"log"
)
func OnPeer(peer p2p.Peer) error {
	fmt.Printf("doing some logic with the peer outside of TCPTransport")
	return nil
}
func main() {
	tcpOpts := p2p.TCPTransportOpts{
		ListenAddr: ":3000",
		Handshake:  p2p.NOPHandshakeFunc,
		Decoder:    p2p.DefaultDecoder{},
		OnPeer:     OnPeer,
	}

	tr := p2p.NewTCPTransport(tcpOpts)
  go func() {
    for{
		msg:= <-tr.Consume()
		fmt.Printf("%v\n", msg)
	}
  }()
	if err := tr.ListenAndAccept(); err != nil {
		log.Fatal(err)
	}
	select {}
}
