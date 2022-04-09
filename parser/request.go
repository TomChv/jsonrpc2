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
// In any case, Request will return a request struct (null or filled) to
// let server returns an identifier if one is found
func Request(body []byte) (*common.Request, error) {
	var req common.Request
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, ErrInvalidBody
	}

	if err := validator.JsonRPCRequest(&req); err != nil {
		return &req, err
	}

	return &req, nil
}
