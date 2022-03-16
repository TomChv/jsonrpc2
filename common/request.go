package common

import "encoding/json"

// RequestParam is a Structured value that holds the parameter values to be used
// during the invocation of the method
type RequestParam = interface{}

// Request is a JSON-RPC 2.0 Object that represent a call to the server
type Request struct {
	// JsonRPC is the version of JSON-RPC protocol, must be "2.0"
	JsonRpc string `json:"jsonrpc" binding:"register"`

	// Method containing the name of the method to be invoked.
	Method string `json:"method" binding:"register"`

	// Params is a Structured value that holds the parameter values to be used
	// during the invocation of the method.
	Params RequestParam `json:"params,omitempty"`

	// ID is an identifier established by the Client that must contain a String,
	// Number, or NULL value if included.
	// If it is not included it is assumed to be a notification.
	ID RequestID `json:"id,omitempty"`
}

func NewRequest() *Request {
	return &Request{
		JsonRpc: JSON_RPC_VERSION,
	}
}

func (r *Request) SetID(id RequestID) *Request {
	r.ID = id

	return r
}

func (r *Request) SetMethod(method string) *Request {
	r.Method = method
	return r
}

func (r *Request) SetParams(params RequestParam) *Request {
	r.Params = params
	return r
}

func (r *Request) Bytes() ([]byte, error) {
	return json.Marshal(r)
}
