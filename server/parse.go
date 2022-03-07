package server

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"

	"github.com/TomChv/jsonrpc2/common"
)

var (
	ErrInvalidJsonVersion    = errors.New("invalid JSON RPC version")
	ErrNoMethodProvided      = errors.New("no method provided")
	ErrInvalidHTTPMethod     = errors.New("http method should be POST")
	ErrInvalidPathRequest    = errors.New("http request should target /")
	ErrInvalidBody           = errors.New("http request invalid body")
	ErrInvalidIdentifierType = errors.New("http request invalid id type")
	ErrInvalidMethodFormat   = errors.New("invalid method format")
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
			// nolint:forcetypeassert
			if req.ID != float64(int(req.ID.(float64))) {
				return nil, InvalidRequestError(ErrInvalidIdentifierType.Error())
			}
			// nolint:forcetypeassert
			req.SetID(int(req.ID.(float64)))
		default:
			return nil, InvalidRequestError(ErrInvalidIdentifierType.Error())
		}
	}

	return &req, nil
}

// procedure is a container around method easily parse which rpc must be called
type procedure struct {
	Service string
	Method  string
}

// parseMethod transform a method string into an object procedure
// If there is no _, it returns the method with no service "" (e.g "sum")
// If there is a _, it split the string and return the first part as a service
// and the second part as a method (e.g "eth_getBalance")
// Otherwise, it returns an error
func parseMethod(method string) (*procedure, *RpcError) {
	toPascalCase := func(str string) string {
		if str == "" {
			return ""
		}
		return strings.ToUpper(string(str[0])) + str[1:]
	}

	switch strings.Count(method, "_") {
	case 0:
		return &procedure{
			Method: toPascalCase(method),
		}, nil
	case 1:
		p := strings.Split(method, "_")
		return &procedure{
			Service: p[0],
			Method:  toPascalCase(p[1]),
		}, nil
	default:
		return nil, InvalidRequestError(ErrInvalidMethodFormat.Error())
	}
}
