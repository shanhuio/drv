package drvapi

// InitInfo contains information of the init procedure.
type InitInfo struct {
	Time              int64  `json:",omitempty"`
	TimeSec           int64  `json:",omitempty"`
	JarvisPassword    string `json:",omitempty"`
	NextcloudPassword string `json:",omitempty"`
}

// InitDoneRequest is the request to set the init info of an endpoint.
type InitDoneRequest struct {
	Name string
	Info *InitInfo
}

// InitDoneResponse is the response of setting the init info of an endpoint.
type InitDoneResponse struct{}
