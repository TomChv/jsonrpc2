package server

import (
	"encoding/json"
	"net/http"

	"github.com/TomChv/jsonrpc2/common"
)

type Response common.Response

// NewResponse create a new Response structure
func NewResponse(id common.RequestID) *Response {
	return &Response{
		JsonRpc: common.JSON_RPC_VERSION,
		ID:      id,
	}
}

func (r *Response) SetID(id common.RequestID) *Response {
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
func (r *Response) SetResult(result common.ResponseResult) *Response {
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
