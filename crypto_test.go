package main

import (
	"bytes"
	//"crypto/des"
	"fmt"
	"testing"
)

func TestCopyEncryptDecrypt(t *testing.T){
	payload:="XYZ fuck u"
	src:=bytes.NewReader([]byte(payload))
	dst:=new(bytes.Buffer)
	key:=newEncryptionKey()
	_, err:=copyEncrypt(key, src, dst)
	 if err !=nil {
		t.Error(err)
	 }
	  fmt.Println(dst.Bytes())
	 out:=new(bytes.Buffer)
	 if _, err:=copyDecrypt(key, dst, out); err !=nil{
		t.Error(err)
	 }
	 if out.String() != payload {
		t.Errorf(" decryption failed")
	 }
      fmt.Println(out.Bytes())
	
}

