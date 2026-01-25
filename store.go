package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"

	"log"
	"os"
	"strings"
)

const defaultfilename = "Shreesh"

type Path struct {
	Pathname string
	Filename string
}

func (p Path) FirstPathName() string {
	paths := strings.Split(p.Pathname, "/")
	if len(paths) == 0 {
		return ""
	}
	return paths[0]
}

func CASPathTransformFunc(key string) Path {
	hash := sha1.Sum([]byte(key))
	hashStr := hex.EncodeToString(hash[:])
	blocksize := 5
	sliceLen := len(hashStr) / blocksize
	paths := make([]string, sliceLen)
	for i := 0; i < sliceLen; i++ {
		paths[i] = hashStr[i*blocksize : (i+1)*blocksize]
	}

	return Path{
		Pathname: strings.Join(paths, "/"),
		Filename: hashStr,
	}
}
func (p Path) FullPath() string {
	return fmt.Sprintf("%s/%s", p.Pathname, p.Filename)
}

type PathTransformFunc func(string) Path

type StoreOpts struct {
	//root id a folder name of the root, containing all the folders /files of the system
	Root              string
	PathTransformFunc PathTransformFunc
}
type Storage struct {
	StoreOpts
}

func (s *Storage) Has(key string) bool {
	pathname := s.PathTransformFunc(key)
	pathNameWithRoot := fmt.Sprintf("%s/%s", s.Root, pathname.FirstPathName())
	_, err := os.Stat(pathNameWithRoot)
	return !os.IsNotExist(err)
}

func NewStore(opts StoreOpts) *Storage {

	if opts.PathTransformFunc == nil {
		opts.PathTransformFunc = DefaultPathTransformFunc
	}

	if len(opts.Root) == 0 {
		opts.Root = defaultfilename
	}
	return &Storage{
		StoreOpts: opts,
	}
}

func (s *Storage) Clear() error {
	return os.RemoveAll(s.Root)
}

func (s *Storage) Delete(key string) error {
	pathname := s.PathTransformFunc(key)

	defer func() {
		log.Printf("deleted [%s] from disk", pathname.FullPath())
	}()
	firstPathwithroot := fmt.Sprintf("%s/%s", s.Root, pathname.FirstPathName())

	return os.RemoveAll(firstPathwithroot)
}

func (s *Storage) Read(key string) (io.Reader, error) {
	f, err := s.readStream(key)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, f)

	return buf, err
}

func (s *Storage) readStream(key string) (io.ReadCloser, error) {
	pathKey := s.PathTransformFunc(key)
	pathKeyWithRoot := fmt.Sprintf("%s/%s", s.Root, pathKey.FullPath())
	return os.Open(pathKeyWithRoot)
}
func (s *Storage) Write(key string, r io.Reader) (int64 , error) {
	return s.writeStream(key, r)

}

func (s *Storage) writeStream(key string, r io.Reader)(int64 , error)  {
	pathname := s.PathTransformFunc(key)
	pathNameWithRoot := fmt.Sprintf("%s/%s", s.Root, pathname.Pathname)
	if err := os.MkdirAll(pathNameWithRoot, os.ModePerm); err != nil {
		return 0, err
	}
	fullPathWithRoot := fmt.Sprintf("%s/%s", s.Root, pathname.FullPath())
	f, err := os.Create(fullPathWithRoot)
	if err != nil {
		return 0,err
	}
	defer f.Close()
	n, err := io.Copy(f, r)
	if err != nil {
		return 0,err
	}
	
	return n,nil
}
