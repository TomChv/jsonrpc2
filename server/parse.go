package server

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/TomChv/jsonrpc2/common"
)

var (
	ErrInvalidJsonVersion = errors.New("invalid JSON RPC version")
	ErrNoMethodProvided   = errors.New("no method provided")
)

// parseRequest transform a HTTP request into a valid Request object.
//
// If the request does not match the specification, it returns a RpcError.
func parseRequest(r *http.Request) (*Request, *RpcError) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, ParsingError(err.Error())
	}

	var req Request
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, InvalidRequestError(err.Error())
	}

	if req.JsonRpc == nil || *req.JsonRpc != common.JSON_RPC_VERSION {
		return nil, InvalidRequestError(ErrInvalidJsonVersion.Error())
	}

	if req.Method == nil {
		return nil, InvalidRequestError(ErrNoMethodProvided.Error())
	}

	// TODO parse ID
	return &req, nil
}
