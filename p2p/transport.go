package p2p

import (
	
	"net"
)

// Peer is an interface which represents the remote node
type Peer interface{
	net.Conn
	Send([]byte) error
}

// Transport is anything that handles communication between nodes in network.
// This can be of the form TCP, UDP and websockets
type Transport interface {
	Dial(adddr string) error
	ListenAndAccept() error
	Consume() <-chan RPC
	Close() error
}