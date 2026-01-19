package p2p


//RPC holds any arbitariry data that is being sent over 
//each transport between two nodes of the network
type RPC struct {
	From    string
	Payload []byte
}
