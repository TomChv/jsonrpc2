package common

import (
	"encoding/json"
	"net/http"
)

// ResponseResult is required on success, this member must not exist if there
// was an error invoking the method
type ResponseResult = interface{}

// Response is a JSON-RPC 2.0 Object that represent a response from the server
type Response struct {
	// JsonRPC is the version of JSON-RPC protocol, must be "2.0"
	JsonRpc string `json:"jsonrpc" binding:"register"`

	// Result is required on success, this member must not exist if there was
	// an error invoking the method
	Result ResponseResult `json:"result,omitempty"`

	// Error is required on error, this member must not exist if there was
	// no error triggered during invocation
	Error *RpcError `json:"error,omitempty"`

	// ID is required, it must be the same value as the id in the Request
	// Object.
	ID RequestID `json:"id"`
}

// NewResponse create a new Response structure
func NewResponse(id RequestID) *Response {
	return &Response{
		JsonRpc: JSON_RPC_VERSION,
		ID:      id,
	}
}

func (r *Response) SetID(id RequestID) *Response {
	r.ID = id
	return r
}

// SetError add error to the Response
//
// NOTE: SetError should not be call if Result is set
func (r *Response) SetError(err *RpcError) *Response {
	r.Error = err
	return r
}

// SetResult add result to the Response
//
// NOTE: SetResult should not be call if RpcError is set
func (r *Response) SetResult(result ResponseResult) *Response {
	r.Result = result

	return r
}

// Send the response to the client
func (r *Response) Send(w http.ResponseWriter) error {
	data, err := json.Marshal(r)
	if err != nil {
		return err
	}

	_, err = w.Write(data)
	if err != nil {
		return err
	}
	return nil
}
