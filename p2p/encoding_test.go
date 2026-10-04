package p2p

import (
	"bytes"
	"encoding/binary"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultDecoderReadsFramedMessagesAndStreams(t *testing.T) {
	var wire bytes.Buffer
	wire.WriteByte(IncomingMessage)
	require.NoError(t, binary.Write(&wire, binary.LittleEndian, uint32(len("payload"))))
	_, err := wire.WriteString("payload")
	require.NoError(t, err)
	wire.WriteByte(IncomingStream)
	_, err = wire.WriteString("file data")
	require.NoError(t, err)

	decoder := DefaultDecoder{}
	var message RPC
	require.NoError(t, decoder.Decode(&wire, &message))
	require.Equal(t, []byte("payload"), message.Payload)

	var stream RPC
	require.NoError(t, decoder.Decode(&wire, &stream))
	require.True(t, stream.Stream)

	remaining, err := io.ReadAll(&wire)
	require.NoError(t, err)
	require.Equal(t, []byte("file data"), remaining)
}
