package server

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/TomChv/jsonrpc2/common"
	"github.com/stretchr/testify/assert"
)

func TestParseRequest(t *testing.T) {
	testCases := []struct {
		name           string
		request        *http.Request
		success        bool
		expectedResult *Request
		expectedError  *RpcError
	}{
		{
			name:           "Invalid request : invalid method",
			request:        httptest.NewRequest("GET", "/", bytes.NewBuffer([]byte(`{"jsonrpc": "2.0", "id": 0, "method": "test"}`))),
			success:        false,
			expectedResult: nil,
			expectedError:  InvalidRequestError(ErrInvalidHTTPMethod.Error()),
		},
		{
			name:           "Invalid request : invalid path",
			request:        httptest.NewRequest("POST", "/unknown", bytes.NewBuffer([]byte(`{"jsonrpc": "2.0", "id": 0, "method": "test"}`))),
			success:        false,
			expectedResult: nil,
			expectedError:  InvalidRequestError(ErrInvalidPathRequest.Error()),
		},
		{
			name:           "Invalid request : missing body",
			request:        httptest.NewRequest("POST", "/", nil),
			success:        false,
			expectedResult: nil,
			expectedError:  InvalidRequestError(ErrInvalidBody),
		},
		{
			name:           "Invalid request : invalid json rpc version",
			request:        httptest.NewRequest("POST", "/", bytes.NewBuffer([]byte(`{"jsonrpc": "3.0", "id": 0, "method": "test"}`))),
			success:        false,
			expectedResult: nil,
			expectedError:  InvalidRequestError(ErrInvalidJsonVersion.Error()),
		},
		{
			name:           "Invalid request : missing method",
			request:        httptest.NewRequest("POST", "/", bytes.NewBuffer([]byte(`{"jsonrpc": "2.0", "id": 0}`))),
			success:        false,
			expectedResult: nil,
			expectedError:  InvalidRequestError(ErrNoMethodProvided.Error()),
		},
		{
			name:           "Invalid request : invalid identifier type - boolean",
			request:        httptest.NewRequest("POST", "/", bytes.NewBuffer([]byte(`{"jsonrpc": "2.0", "method": "/test", "id": true}`))),
			success:        false,
			expectedResult: nil,
			expectedError:  InvalidRequestError(ErrInvalidIdentifierType.Error()),
		},
		{
			name:           "Invalid request : invalid identifier type - struct",
			request:        httptest.NewRequest("POST", "/", bytes.NewBuffer([]byte(`{"jsonrpc": "2.0", "method": "/test", "id": { "foo": "bar" }}`))),
			success:        false,
			expectedResult: nil,
			expectedError:  InvalidRequestError(ErrInvalidIdentifierType.Error()),
		},
		{
			name:           "Valid request : only method",
			request:        httptest.NewRequest("POST", "/", bytes.NewBuffer([]byte(`{"jsonrpc": "2.0", "method": "/test"}`))),
			success:        true,
			expectedResult: common.NewRequest().SetMethod("/test"),
			expectedError:  nil,
		},

		{
			name:           "Valid request : string identifier",
			request:        httptest.NewRequest("POST", "/", bytes.NewBuffer([]byte(`{"jsonrpc": "2.0", "method": "/test", "id": "fake_id"}`))),
			success:        true,
			expectedResult: common.NewRequest().SetMethod("/test").SetID("fake_id"),
			expectedError:  nil,
		},
		{
			name:           "Valid request : number identifier",
			request:        httptest.NewRequest("POST", "/", bytes.NewBuffer([]byte(`{"jsonrpc": "2.0", "method": "/test", "id": 4}`))),
			success:        true,
			expectedResult: common.NewRequest().SetMethod("/test").SetID(4),
			expectedError:  nil,
		},
		{
			name:           "Valid request : null identifier",
			request:        httptest.NewRequest("POST", "/", bytes.NewBuffer([]byte(`{"jsonrpc": "2.0", "method": "/test", "id": null}`))),
			success:        true,
			expectedResult: common.NewRequest().SetMethod("/test").SetID(nil),
			expectedError:  nil,
		},
		{
			name:           "Valid request : with param string",
			request:        httptest.NewRequest("POST", "/", bytes.NewBuffer([]byte(`{"jsonrpc": "2.0", "method": "/test", "params": "test"}`))),
			success:        true,
			expectedResult: common.NewRequest().SetMethod("/test").SetParams("test"),
			expectedError:  nil,
		},
		{
			name:           "Valid request : with param number",
			request:        httptest.NewRequest("POST", "/", bytes.NewBuffer([]byte(`{"jsonrpc": "2.0", "method": "/test", "params": 4}`))),
			success:        true,
			expectedResult: common.NewRequest().SetMethod("/test").SetParams(float64(4)),
			expectedError:  nil,
		},
		{
			name:    "Valid request : with param struct",
			request: httptest.NewRequest("POST", "/", bytes.NewBuffer([]byte(`{"jsonrpc": "2.0", "method": "/test", "params": {"foo": "bar", "baz": 4 }}`))),
			success: true,
			expectedResult: common.NewRequest().SetMethod("/test").SetParams(map[string]interface{}{
				"foo": "bar",
				"baz": float64(4),
			}),
			expectedError: nil,
		},
		{
			name:    "Valid request : with param nested struct",
			request: httptest.NewRequest("POST", "/", bytes.NewBuffer([]byte(`{"jsonrpc": "2.0", "method": "/test", "params": {"foo": "bar", "baz": 4, "fizz": { "bool": true }}}`))),
			success: true,
			expectedResult: common.NewRequest().SetMethod("/test").SetParams(map[string]interface{}{
				"foo": "bar",
				"baz": float64(4),
				"fizz": map[string]interface{}{
					"bool": true,
				},
			}),
			expectedError: nil,
		},
		{
			name:    "Valid request : with param nested struct with identifier",
			request: httptest.NewRequest("POST", "/", bytes.NewBuffer([]byte(`{"jsonrpc": "2.0", "method": "/test", "id": "fake_id", "params": {"foo": "bar", "baz": 4, "fizz": { "bool": true }}}`))),
			success: true,
			expectedResult: common.NewRequest().SetID("fake_id").SetMethod("/test").SetParams(map[string]interface{}{
				"foo": "bar",
				"baz": float64(4),
				"fizz": map[string]interface{}{
					"bool": true,
				},
			}),
			expectedError: nil,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			res, err := parseRequest(tt.request)

			assert.Equal(t, tt.expectedResult, res)
			assert.Equal(t, tt.expectedError, err)
		})
	}
}

func TestParseMethod(t *testing.T) {
	testCases := []struct {
		name           string
		method         string
		success        bool
		expectedResult *procedure
		expectedError  *RpcError
	}{
		{
			name:           "Invalid method : 2 underscore",
			method:         "foo_bar_baz",
			success:        false,
			expectedResult: nil,
			expectedError:  InvalidRequestError(ErrInvalidMethodFormat.Error()),
		},
		{
			name:           "Valid method : no underscore",
			method:         "foo",
			success:        true,
			expectedResult: &procedure{Service: "", Method: "Foo"},
			expectedError:  nil,
		},
		{
			name:           "Valid method : no underscore and already PascalCase",
			method:         "Foo",
			success:        true,
			expectedResult: &procedure{Service: "", Method: "Foo"},
			expectedError:  nil,
		},
		{
			name:           "Valid method : underscore",
			method:         "eth_getBalance",
			success:        true,
			expectedResult: &procedure{Service: "eth", Method: "GetBalance"},
			expectedError:  nil,
		},
		{
			name:           "Valid method : underscore with already PascalCase",
			method:         "eth_GetBalance",
			success:        true,
			expectedResult: &procedure{Service: "eth", Method: "GetBalance"},
			expectedError:  nil,
		},
		{
			name:           "Valid method : underscore with empty method",
			method:         "eth_",
			success:        true,
			expectedResult: &procedure{Service: "eth", Method: ""},
			expectedError:  nil,
		},
		{
			name:           "Valid method : underscore with empty method",
			method:         "_getBalance",
			success:        true,
			expectedResult: &procedure{Service: "", Method: "GetBalance"},
			expectedError:  nil,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			res, err := parseMethod(tt.method)

			assert.Equal(t, tt.expectedResult, res)
			assert.Equal(t, tt.expectedError, err)
		})
	}
}

func TestParseParams(t *testing.T) {
	testCases := []struct {
		name           string
		success        bool
		args           []reflect.Type
		params         interface{}
		expectedResult []interface{}
		expectedError  *RpcError
	}{
		{
			name:           "Valid : no arguments",
			success:        true,
			args:           []reflect.Type{},
			params:         nil,
			expectedResult: []interface{}{},
			expectedError:  nil,
		},
		{
			name:           "Valid : parse one arg : string",
			success:        true,
			args:           []reflect.Type{reflect.TypeOf("")},
			params:         "foo",
			expectedResult: []interface{}{"foo"},
			expectedError:  nil,
		},
		{
			name:           "Valid : parse one arg : int",
			success:        true,
			args:           []reflect.Type{reflect.TypeOf(4)},
			params:         4,
			expectedResult: []interface{}{4},
			expectedError:  nil,
		},
		{
			name:           "Valid : parse one arg : boolean",
			success:        true,
			args:           []reflect.Type{reflect.TypeOf(true)},
			params:         false,
			expectedResult: []interface{}{false},
			expectedError:  nil,
		},
		{
			name:           "Valid : parse one arg : float",
			success:        true,
			args:           []reflect.Type{reflect.TypeOf(float64(2))},
			params:         float64(2),
			expectedResult: []interface{}{float64(2)},
			expectedError:  nil,
		},
		{
			name:           "Valid : parse one arg : object",
			success:        true,
			args:           []reflect.Type{reflect.TypeOf(struct{ Foo string }{Foo: ""})},
			params:         struct{ Foo string }{Foo: "foo"},
			expectedResult: []interface{}{struct{ Foo string }{Foo: "foo"}},
			expectedError:  nil,
		},
		{
			name:           "Valid : parse one arg : array",
			success:        true,
			args:           []reflect.Type{reflect.TypeOf([]int{0})},
			params:         []int{1, 2, 3},
			expectedResult: []interface{}{[]int{1, 2, 3}},
			expectedError:  nil,
		},
		{
			name:           "Valid : parse multi arg : int",
			success:        true,
			args:           []reflect.Type{reflect.TypeOf(0), reflect.TypeOf(0)},
			params:         []int{1, 2},
			expectedResult: []interface{}{1, 2},
			expectedError:  nil,
		},
		{
			name:           "Valid : parse multi arg : string",
			success:        true,
			args:           []reflect.Type{reflect.TypeOf(""), reflect.TypeOf("")},
			params:         []string{"foo", "bar"},
			expectedResult: []interface{}{"foo", "bar"},
			expectedError:  nil,
		},
		{
			name:           "Valid : parse multi arg : boolean",
			success:        true,
			args:           []reflect.Type{reflect.TypeOf(false), reflect.TypeOf(false)},
			params:         []bool{false, true},
			expectedResult: []interface{}{false, true},
			expectedError:  nil,
		},
		{
			name:           "Valid : parse multi arg : mix primitive type",
			success:        true,
			args:           []reflect.Type{reflect.TypeOf(false), reflect.TypeOf(""), reflect.TypeOf(0)},
			params:         []interface{}{true, "foo", 5},
			expectedResult: []interface{}{true, "foo", 5},
			expectedError:  nil,
		},
		{
			name:           "Valid : parse multi arg : mix primitive type with array",
			success:        true,
			args:           []reflect.Type{reflect.TypeOf(false), reflect.TypeOf(""), reflect.TypeOf([]int{0})},
			params:         []interface{}{true, "foo", []int{1, 2, 3}},
			expectedResult: []interface{}{true, "foo", []int{1, 2, 3}},
			expectedError:  nil,
		},
		{
			name:           "Valid : parse multi arg : mix primitive type with array and object",
			success:        true,
			args:           []reflect.Type{reflect.TypeOf(false), reflect.TypeOf(""), reflect.TypeOf([]int{0}), reflect.TypeOf(struct{ Foo string }{Foo: ""})},
			params:         []interface{}{true, "foo", []int{1, 2, 3}, struct{ Foo string }{Foo: "foo"}},
			expectedResult: []interface{}{true, "foo", []int{1, 2, 3}, struct{ Foo string }{Foo: "foo"}},
			expectedError:  nil,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			res, err := parseParams(tt.args, tt.params)

			assert.Equal(t, tt.expectedResult, res)
			assert.Equal(t, tt.expectedError, err)
		})
	}
}
