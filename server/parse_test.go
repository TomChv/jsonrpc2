package server

import (
	"bytes"
	"net/http"
	"net/http/httptest"
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
