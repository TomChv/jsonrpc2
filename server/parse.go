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
	ErrInvalidJsonVersion      = errors.New("invalid JSON RPC version")
	ErrNoMethodProvided        = errors.New("no method provided")
	ErrInvalidHTTPMethod       = errors.New("http method should be POST")
	ErrInvalidPathRequest      = errors.New("http request should target /")
	ErrInvalidBody             = errors.New("http request invalid body")
	ErrInvalidIdentifierType   = errors.New("http request invalid id type")
	ErrInvalidMethodFormat     = errors.New("invalid method format")
	ErrInvalidArgsNumber       = errors.New("invalid number of arguments")
	ErrInvalidArgExpectedSlice = errors.New("invalid argument, expected slice")
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

// parseParams convert any param into types send in args
//   If no args         -> return empty
//   If 1 arg           -> directly parse the param and return it a single
//   IF 2 or more arg   -> verify that param is an array and loop through it to
//  convert it to an array of interface with correct type
func parseParams(args []reflect.Type, param interface{}) ([]interface{}, *RpcError) {
	switch len(args) {
	case 0:
		return []interface{}{}, nil
	case 1:
		p, err := parseParam(args[0], param)
		return []interface{}{p}, err
	default:
		if reflect.TypeOf(param).Kind() != reflect.Slice {
			return nil, InvalidParamsError(ErrInvalidArgsNumber.Error())
		}

		params, err := convertInterfaceToArray(param)
		if err != nil {
			return nil, err
		}

		res := make([]interface{}, len(args))
		for i, e := range params {
			p, err := parseParam(args[i], e)
			if err != nil {
				return nil, err
			}
			res[i] = p
		}

		return res, nil
	}
}

// parseParam convert the param into the type of the arg
// Since a simple reflect is not enough to verify if the param is type of arg
// this function use json.Unmarshal to correctly convert the param
// FIXME: this function is more a hack than a real way to convert type
func parseParam(arg reflect.Type, param interface{}) (interface{}, *RpcError) {
	expectedType := reflect.StructOf([]reflect.StructField{{
		Name: "Placeholder",
		Type: arg,
	}})
	expected := reflect.New(expectedType).Interface()

	placeholder := struct {
		Placeholder interface{}
	}{
		Placeholder: param,
	}

	data, err := json.Marshal(placeholder)
	if err != nil {
		return nil, InvalidParamsError(err.Error())
	}

	err = json.Unmarshal(data, &expected)
	if err != nil {
		return nil, InvalidParamsError(err.Error())
	}

	value := reflect.ValueOf(expected).Elem().FieldByName("Placeholder")
	return value.Interface(), nil
}

// convertInterfaceToArray is a utility function used to transform
// an interface into an array of interface
// The result can then be used to populate arguments to the dispatcher
func convertInterfaceToArray(value interface{}) ([]interface{}, *RpcError) {
	var out []interface{}

	reflectValue := reflect.ValueOf(value)
	if reflectValue.Kind() != reflect.Slice {
		return nil, InvalidParamsError(ErrInvalidArgExpectedSlice.Error())
	}

	for i := 0; i < reflectValue.Len(); i++ {
		out = append(out, reflectValue.Index(i).Interface())
	}

	return out, nil
}
