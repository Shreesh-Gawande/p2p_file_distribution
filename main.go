package main

import (

	//	"bytes"
	"bytes"
	"file_distribution_system/p2p"
	"fmt"

	"log"
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
	folderName := listenAddr[1:]
	fileTransport := FileServerOpts{
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
	s1 := makeserver(":3000", "")
	s2 := makeserver(":4000", ":3000")
	go func() {
		log.Fatal(s1.Start())
	}()
	time.Sleep(2 * time.Second)
	go s2.Start()

	time.Sleep(5 * time.Second)
	for i:=0 ; i<10 ; i++ {
     	data := bytes.NewReader([]byte("my big data file is here"))
	    s2.Store(fmt.Sprintf("myprivateData %d",i), data)
		time.Sleep(time.Millisecond*5)
	}


	/* 	r, err:=s2.Get("myprivateData")
	if err!=nil{
		log.Fatal(err)
	}
	b, err:= ioutil.ReadAll(r)
	if err !=nil{
		log.Fatal(err)
	}

	fmt.Println(string(b))  */

	select {}
}
