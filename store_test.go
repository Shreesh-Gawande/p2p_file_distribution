package main

import (
	"bytes"
	
	"testing"
)

func TestPathTransformFunc(t *testing.T){
	key:="rishi ke nudes"
	pathKey:=CASPathTransformFunc(key)
	expectedOrignalKey:="ca96112975c910d49bac6fa7decce83882e95f5c"
	expectedPathname:="ca961/12975/c910d/49bac/6fa7d/ecce8/3882e/95f5c"
	if pathKey.Pathname !=expectedPathname{
		t.Errorf("have %s want %s", pathKey.Pathname, expectedPathname)
	}
	if pathKey.Filename !=expectedOrignalKey{
		t.Errorf("have %s want %s", pathKey.Filename, expectedOrignalKey)
	}
	
}

func TestStore(t *testing.T) {
    opts:=StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}
	s:=NewStore(opts)
     data:=bytes.NewReader([]byte("some jpg bytes"))
	if err:=s.writeStream("myspecialpicture", data); err!=nil{
		t.Error(err)
	} 
}