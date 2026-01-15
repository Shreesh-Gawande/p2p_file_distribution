package main

import (
	"bytes"
	"fmt"
	"io"

	"testing"
)

func TestPathTransformFunc(t *testing.T) {
	key := "rishi ke nudes"
	pathKey := CASPathTransformFunc(key)
	expectedOrignalKey := "ca96112975c910d49bac6fa7decce83882e95f5c"
	expectedPathname := "ca961/12975/c910d/49bac/6fa7d/ecce8/3882e/95f5c"
	if pathKey.Pathname != expectedPathname {
		t.Errorf("have %s want %s", pathKey.Pathname, expectedPathname)
	}
	if pathKey.Filename != expectedOrignalKey {
		t.Errorf("have %s want %s", pathKey.Filename, expectedOrignalKey)
	}

}
func TestStoreDelete(t *testing.T) {
	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}
	s := NewStore(opts)
	key := "myspecialpicture"

	data := []byte("some jpg bytes")
	if err := s.writeStream("myspecialpicture", bytes.NewReader(data)); err != nil {
		t.Error(err)
	}

	if err := s.Delete(key); err != nil {
		t.Error(err)
	}
}

func TestStore(t *testing.T) {
	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}
	s := NewStore(opts)
	key := "myspecialpicture"

	data := []byte("some jpg bytes")
	if err := s.writeStream("myspecialpicture", bytes.NewReader(data)); err != nil {
		t.Error(err)
	}

	if ok := s.Has(key); !ok {
		t.Errorf("expected to have key %s", key)
	}

	r, err := s.Read(key)
	if err != nil {
		t.Error(err)
	}
	b, _ := io.ReadAll(r)
	fmt.Println(string(b))

	if string(b) != string(data) {
		t.Errorf("want %s have %s", string(data), string(b))
	}

	

	s.Delete(key)

}
