package p2p

// Peer is an interface which represents the remote node
type Peer interface{}

// Transport is anything that handles communication between nodes in network.
// This can be of the form TCP, UDP and websockets
type Transport interface {
	ListenAndAccept() error
}