package main

import (
	"bytes"
	"file_distribution_system/p2p"
	"fmt"
	"io/ioutil"

	//o/ioutil"
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
		EncKey: newEncryptionKey(),
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
	s3:=makeserver(":5000", ":4000", ":3000")
	go func() {
		log.Fatal(s1.Start())
		time.Sleep(1 * time.Second)
		//log.Fatal(s2.Start())
	}()
		go func() {
		//log.Fatal(s1.Start())
		//time.Sleep(1 * time.Second)
		log.Fatal(s2.Start())
	}()
	time.Sleep(2 * time.Second)
	go s3.Start()

	time.Sleep(5 * time.Second)

	for i:=0; i<20 ; i++ {
		key:=fmt.Sprintf("pictured.png(%d)",i)
		data := bytes.NewReader([]byte("my big data file is here"))
	    s3.Store(key, data) 
		
		if err:=s3.store.Delete(s3.ID,key) ; err!=nil{
			log.Fatal(err)
		}
		r, err:=s3.Get(key)
	if err!=nil{
		log.Fatal(err)
	}
	b, err:= ioutil.ReadAll(r)
	if err !=nil{
		log.Fatal(err)
	}

	fmt.Println(string(b)) 

	}
    
  


	

		 
	 

 

}
