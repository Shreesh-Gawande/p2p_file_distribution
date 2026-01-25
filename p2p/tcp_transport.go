package p2p

import (
	"errors"
	"fmt"
	"net"
	"sync"
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
		rpcch:            make(chan RPC,1024),
	}
}
//Adder impliments the Transport interface return the address
//the transport is accepting connections

func(t *TCPTransport) Addr() string{
	return t.ListenAddr
}


// Consume implements transport interface which will return read only channel
// for reading the incomming message recieved from another peer in the network
func (t *TCPTransport) Consume() <-chan RPC {
	return t.rpcch
}

// Close implements the Transport interface
func (t *TCPTransport) Close() error {
	return t.listener.Close()
}

func (t *TCPTransport) Dial(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}

	go t.handleConn(conn, true)
	return nil
}

// TCPPeer represents the remote node over a TCP established connection
type TCPPeer struct {
	//The underlying connection of the peer. Which in this case is a TCP connection
	net.Conn
	//if we make the connection outbound true else false
	outbound bool

	wg *sync.WaitGroup
}

func(p *TCPPeer) CloseStream() {
	p.wg.Done()
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		Conn:     conn,
		outbound: outbound,
		wg:       &sync.WaitGroup{},
	}
}

func (p *TCPPeer) Send(data []byte) error {
	_, err := p.Conn.Write(data)
	return err
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
		if errors.Is(err, net.ErrClosed) {
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
	defer func() {
		fmt.Printf("dropping the connection %s", err)
		conn.Close()
	}()
	err = t.Handshake(peer)
	if err != nil {
		return
	}
	if t.OnPeer != nil {
		if err = t.OnPeer(peer); err != nil {
			return
		}
	}

	
	for {
		rpc := RPC{}
		err := t.Decoder.Decode(conn, &rpc)
		if err != nil {
			return
		}
		rpc.From = conn.RemoteAddr().String()
		if rpc.Stream {
			peer.wg.Add(1)
			fmt.Printf("[%s] incomming stream, waiting ...\n", conn.RemoteAddr())
			peer.wg.Wait()
			fmt.Printf("[%s] stream closed, resuming read loop \n", conn.RemoteAddr())
			continue

		}

		
		t.rpcch <- rpc

		fmt.Println("stream done continueing normal reed loop")
	}
}
