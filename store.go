package main

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)
type Path struct{
	Pathname string
	Filename string
}

func CASPathTransformFunc(key string) Path{
	hash:=sha1.Sum([]byte(key))
	hashStr:=hex.EncodeToString(hash[:])
	blocksize:=5
	sliceLen:=len(hashStr)/blocksize
	paths:=make([]string,sliceLen)
	for i := 0; i < sliceLen; i++ {
		paths[i]=hashStr[i*blocksize:(i+1)*blocksize]
	}

	return Path{
		Pathname: strings.Join(paths,"/"),
		Filename: hashStr,
	}
}
func(p Path) FullPath() string{
	return fmt.Sprintf("%s/%s",p.Pathname,p.Filename)
}

type PathTransformFunc func(string) Path

type StoreOpts struct {
	PathTransformFunc PathTransformFunc
}
type Storage struct {
	StoreOpts
}


func NewStore(opts StoreOpts) *Storage {
	return &Storage{
		StoreOpts: opts,
	}
}



func (s *Storage) writeStream(key string, r io.Reader) error{
	pathname:= s.PathTransformFunc(key)
	if err:=os.MkdirAll(pathname.Pathname, os.ModePerm);err !=nil{
		return err
	}

	fullPath:= pathname.FullPath()
	f, err:=os.Create(fullPath)
	if err!=nil {
		return err
	}
	n, err:=io.Copy(f,r)
	if err !=nil{
		return err
	}
	log.Printf("written(%d) bytes to disc %s ",n,fullPath)
	return nil
}
