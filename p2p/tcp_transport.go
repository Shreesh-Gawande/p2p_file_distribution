package p2p

import (
	"fmt"
	"net"
	"sync"
)

type TCPTransportOpts struct {
	ListenAddr string
	Handshake  HandShakeFunc
	Decoder    Decoder
}

type TCPTransport struct {
	TCPTransportOpts
	listener net.Listener

	mu    sync.RWMutex
	peers map[net.Addr]Peer
}

// TCPPeer represents the remote node over a TCP established connection
type TCPPeer struct {
	//conn is the underlying connection of the peer
	conn net.Conn
	//if we make the connection outbound true else false
	outbound bool
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		conn:     conn,
		outbound: outbound,
	}
}

func NewTCPTransport(opts TCPTransportOpts) *TCPTransport {
	return &TCPTransport{
		TCPTransportOpts: opts,
	}
}

func (t *TCPTransport) ListenAndAccept() error {
	var err error
	t.listener, err = net.Listen("tcp", t.ListenAddr)
	if err != nil {
		return err
	}
	go t.startAcceptLoop()
	return nil

}

func (t *TCPTransport) startAcceptLoop() {
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			fmt.Printf("TCP accept error: %s\n", err)
			continue
		}
		fmt.Printf("new incoming connection %v \n", conn)

		go t.handleConn(conn)
	}
}

func (t *TCPTransport) handleConn(conn net.Conn) {
	peer := NewTCPPeer(conn, true)
	err := t.Handshake(peer)
	if err != nil {
		conn.Close()
		fmt.Printf("handshake failed %v\n", err)
		return
	}

	msg := &Message{}
	for {
		if err := t.Decoder.Decode(conn, msg); err != nil {
			fmt.Printf("TCP error: %s \n", &err)
			continue
		}

		msg.From=conn.RemoteAddr()
		fmt.Printf(" message: %v \n", msg)
	}
}
