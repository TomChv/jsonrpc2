package client

import (
	"encoding/json"

	"github.com/TomChv/jsonrpc2/common"
)

// Request is an extension of type common.Request with client's method
type Request common.Request

func NewRequest() *Request {
	return &Request{
		JsonRpc: common.JSON_RPC_VERSION,
	}
}

func (r *Request) SetID(id common.RequestID) *Request {
	r.ID = id

	return r
}

func (r *Request) SetMethod(method string) *Request {
	r.Method = method
	return r
}

func (r *Request) SetParams(params common.RequestParam) *Request {
	r.Params = params
	return r
}

func (r *Request) Bytes() ([]byte, error) {
	return json.Marshal(r)
}
