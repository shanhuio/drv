package drvapi

// RegisterRequest is the request for setting up an endpoint's public key using
// a passcode.
type RegisterRequest struct {
	Name       string
	PassCode   string
	ControlKey string `json:",omitempty"`
}

// RegisterDoorwayRequest is the request for registering a doorway fabrics
// connection.
type RegisterDoorwayRequest struct {
	PublicKey string `json:",omitempty"`
}

// RegisterTunnelRequest is the request for registering a hometunn fabrics
// connection.
type RegisterTunnelRequest struct {
	Name      string
	PassCode  string
	PublicKey string
}
