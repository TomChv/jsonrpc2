package server

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"reflect"

	"github.com/TomChv/jsonrpc2/common"
)

var (
	ErrInvalidJsonVersion    = errors.New("invalid JSON RPC version")
	ErrNoMethodProvided      = errors.New("no method provided")
	ErrInvalidHTTPMethod     = errors.New("http method should be POST")
	ErrInvalidPathRequest    = errors.New("http request should target /")
	ErrInvalidBody           = errors.New("http request invalid body")
	ErrInvalidIdentifierType = errors.New("http request invalid id type")
)

// parseRequest transform a HTTP request into a valid Request object.
//
// If the request does not match the specification, it returns a RpcError.
func parseRequest(r *http.Request) (*Request, *RpcError) {
	if r.Method != http.MethodPost {
		return nil, InvalidRequestError(ErrInvalidHTTPMethod.Error())
	}

	if r.URL.Path != "/" {
		return nil, InvalidRequestError(ErrInvalidPathRequest.Error())
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, ParsingError(err.Error())
	}

	var req Request
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, InvalidRequestError(ErrInvalidBody)
	}

	if req.JsonRpc != common.JSON_RPC_VERSION {
		return nil, InvalidRequestError(ErrInvalidJsonVersion.Error())
	}

	if req.Method == "" {
		return nil, InvalidRequestError(ErrNoMethodProvided.Error())
	}

	if req.ID != nil {
		switch reflect.TypeOf(req.ID).String() {
		case "string":
			break
		case "float64":
			// Verify if it's an integer or a float
			if req.ID != float64(int(req.ID.(float64))) {
				return nil, InvalidRequestError(ErrInvalidIdentifierType.Error())
			}
			req.SetID(int(req.ID.(float64)))
			break
		default:
			return nil, InvalidRequestError(ErrInvalidIdentifierType.Error())
		}
	}

	return &req, nil
}
