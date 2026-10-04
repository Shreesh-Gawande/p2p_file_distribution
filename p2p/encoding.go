package p2p

import (
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"
)

type Decoder interface {
	Decode(io.Reader, *RPC) error
}

type GOBDecoder struct{}

func (dec GOBDecoder) Decode(r io.Reader, msg *RPC) error {
	return gob.NewDecoder(r).Decode(msg)
}

type DefaultDecoder struct{}

func (dec DefaultDecoder) Decode(r io.Reader, msg *RPC) error {
	var kind [1]byte
	if _, err := io.ReadFull(r, kind[:]); err != nil {
		return err
	}

	switch kind[0] {
	case IncomingStream:
		msg.Stream = true
		return nil
	case IncomingMessage:
		var size [4]byte
		if _, err := io.ReadFull(r, size[:]); err != nil {
			return err
		}
		payloadSize := binary.LittleEndian.Uint32(size[:])
		if payloadSize > 16<<20 {
			return fmt.Errorf("message payload is too large: %d bytes", payloadSize)
		}
		msg.Payload = make([]byte, payloadSize)
		_, err := io.ReadFull(r, msg.Payload)
		return err
	default:
		return fmt.Errorf("unknown message type: %d", kind[0])
	}
}
