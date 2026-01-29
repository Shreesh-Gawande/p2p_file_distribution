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

func (s *Storage) Has(id string,key string) bool {
	pathname := s.PathTransformFunc(key)
	pathNameWithRoot := fmt.Sprintf("%s/%s/%s", s.Root,id, pathname.FirstPathName())
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

func (s *Storage) Delete(id string,key string) error {
	pathname := s.PathTransformFunc(key)

	defer func() {
		log.Printf("deleted [%s] from disk", pathname.FullPath())
	}()
	firstPathwithroot := fmt.Sprintf("%s/%s/%s", s.Root,id, pathname.FirstPathName())

	return os.RemoveAll(firstPathwithroot)
}

func (s *Storage) Read(id string,key string) (int64 ,io.Reader, error) {
	return s.readStream(id,key)
}

func (s *Storage) readStream(id string,key string) ( int64,io.ReadCloser, error) {
	pathKey := s.PathTransformFunc(key)
	pathKeyWithRoot := fmt.Sprintf("%s/%s/%s", s.Root,id, pathKey.FullPath())
	file, err:= os.Open(pathKeyWithRoot)
	if err!=nil{
		return 0, nil, err
	}
	n, err :=file.Stat()
	if err !=nil{
		return 0, nil, err
	}
	return n.Size(),file, nil
}
func (s *Storage) Write(id string,key string, r io.Reader) (int64 , error) {
	return s.writeStream(id,key, r)

}
func(s *Storage) WriteDecrypt(id string,encKey []byte, key string, r io.Reader)(int64, error){
    f, err:=s.openFileForWriting(id,key)
	if err != nil {
		return 0,err
	}
	defer f.Close()
	n, err := copyDecrypt(encKey,r,f)

	
	return int64(n),nil
}

func(s *Storage) openFileForWriting(id string,key string)(*os.File, error){
	pathname := s.PathTransformFunc(key)
	pathNameWithRoot := fmt.Sprintf("%s/%s/%s", s.Root,id, pathname.Pathname)
	if err := os.MkdirAll(pathNameWithRoot, os.ModePerm); err != nil {
		return nil, err
	}
	fullPathWithRoot := fmt.Sprintf("%s/%s/%s", s.Root,id, pathname.FullPath())
	 return  os.Create(fullPathWithRoot)
}

func (s *Storage) writeStream(id string,key string, r io.Reader)(int64 , error)  {
    f, err:=s.openFileForWriting(id,key)
	if err != nil {
		return 0,err
	}
	defer f.Close()
	return  io.Copy(f, r)
	
}
