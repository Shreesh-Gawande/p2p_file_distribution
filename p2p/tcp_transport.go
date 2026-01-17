package p2p

import (
	"errors"
	"fmt"
	"net"
)

type TCPTransportOpts struct {
	ListenAddr string
	Handshake  HandShakeFunc
	Decoder    Decoder
	OnPeer     func(Peer) error
}

type TCPTransport struct {
	TCPTransportOpts
	listener net.Listener
	rpcch    chan RPC


}

func NewTCPTransport(opts TCPTransportOpts) *TCPTransport {
	return &TCPTransport{
		TCPTransportOpts: opts,
		rpcch:            make(chan RPC),
	}
}

// Consume implements transport interface which will return read only channel
// for reading the incomming message recieved from another peer in the network
func (t *TCPTransport) Consume() <-chan RPC {
	return t.rpcch
}

//Close implements the Transport interface
func (t *TCPTransport) Close() error{
	return t.listener.Close()
}


func (t *TCPTransport) Dial(addr string) error{
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	go t.handleConn(conn,true)
	return nil
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

func (p *TCPPeer) Send(data [] byte) error {
	_ , err:=p.conn.Write(data)
	return err
}

// REmoteAddr implimets peer interface and return the remote address of the peer
func (p *TCPPeer) RemoteAddr() net.Addr{
	return p.conn.RemoteAddr()
}

// Close implements Peer interface to close the underlying connection
func (p *TCPPeer) Close() error {
	return p.conn.Close()
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
		if errors.Is(err, net.ErrClosed){
			return
		}
		if err != nil {
			fmt.Printf("TCP accept error: %s\n", err)
			continue
		}
		fmt.Printf("new incoming connection %v \n", conn)

		go t.handleConn(conn, false)
	}
}

func (t *TCPTransport) handleConn(conn net.Conn, outbound bool) {
	var err error
	peer := NewTCPPeer(conn, outbound)
	defer func ()  {
	    fmt.Printf("dropping the connection %s", err)
		conn.Close()	
	}()
	err = t.Handshake(peer)
	if err != nil {
		return
	}
  if t.OnPeer !=nil {
	if err=t.OnPeer(peer); err!=nil{
		return 
	}
  }

	rpc := RPC{}
	for {
		if err := t.Decoder.Decode(conn, &rpc); err != nil {
			fmt.Printf("TCP error: %v \n", &err)
			return
		}

		rpc.From = conn.RemoteAddr()
		t.rpcch <- rpc
	}
}
