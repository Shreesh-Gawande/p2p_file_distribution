package main

import (
	"file_distribution_system/p2p"

	"log"
)

func main() {

	tcpOpts := p2p.TCPTransportOpts{
		ListenAddr: ":3000",
		Handshake:  p2p.NOPHandshakeFunc,
		Decoder:    p2p.DefaultDecoder{},
	}

	tr := p2p.NewTCPTransport(tcpOpts)

	if err := tr.ListenAndAccept(); err != nil {
		log.Fatal(err)
	}
	select {}
}
