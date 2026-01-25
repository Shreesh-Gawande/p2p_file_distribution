package p2p
const (
	IncomingMessage = 0x1
	IncomingStream  = 0x2
)

//RPC holds any arbitariry data that is being sent over 
//each transport between two nodes of the network
type RPC struct {
	From    string
	Payload []byte
	Stream bool
}
