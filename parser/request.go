package parser

import (
	"encoding/json"
	"errors"

	"github.com/TomChv/jsonrpc2/common"
	"github.com/TomChv/jsonrpc2/validator"
)

var (
	ErrInvalidBody = errors.New("http request invalid body")
)

// Request convert an array of byte into a valid Request object.
//
// If the request does not match JSON RPC specification, it returns
// an error
// TODO(TomChv): Fix #8
func Request(body []byte) (*common.Request, error) {
	var req common.Request
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, ErrInvalidBody
	}

	if err := validator.JsonRPCRequest(&req); err != nil {
		return nil, err
	}

	return &req, nil
}
