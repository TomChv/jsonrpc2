package parser

import (
	"testing"

	"github.com/TomChv/jsonrpc2/common"
	"github.com/TomChv/jsonrpc2/validator"
	"github.com/stretchr/testify/assert"
)

func TestRequest(t *testing.T) {
	testCases := []struct {
		name           string
		body           []byte
		success        bool
		expectedResult *common.Request
		expectedError  error
	}{
		{
			name:           "Missing body",
			success:        false,
			expectedResult: nil,
			expectedError:  ErrInvalidBody,
		},
		{
			name:           "Invalid json rpc version",
			body:           []byte(`{"jsonrpc": "3.0", "id": 0, "method": "test"}`),
			success:        false,
			expectedResult: &common.Request{JsonRpc: "3.0", Method: "test", ID: 0},
			expectedError:  validator.ErrInvalidJsonVersion,
		},
		{
			name:           "Missing method",
			body:           []byte(`{"jsonrpc": "2.0", "id": 0}`),
			success:        false,
			expectedResult: &common.Request{JsonRpc: "2.0", ID: 0},
			expectedError:  validator.ErrMissingMethod,
		},
		{
			name:           "Invalid identifier type - boolean",
			body:           []byte(`{"jsonrpc": "2.0", "method": "/test", "id": true}`),
			success:        false,
			expectedResult: &common.Request{JsonRpc: "2.0", Method: "/test"},
			expectedError:  validator.ErrInvalidIdentifierType,
		},
		{
			name:           "Invalid identifier type - struct",
			body:           []byte(`{"jsonrpc": "2.0", "method": "/test", "id": { "foo": "bar" }}`),
			success:        false,
			expectedResult: &common.Request{JsonRpc: "2.0", Method: "/test"},
			expectedError:  validator.ErrInvalidIdentifierType,
		},
		{
			name:           "Only method",
			body:           []byte(`{"jsonrpc": "2.0", "method": "/test"}`),
			success:        true,
			expectedResult: common.NewRequest().SetMethod("/test"),
			expectedError:  nil,
		},

		{
			name:           "String identifier",
			body:           []byte(`{"jsonrpc": "2.0", "method": "/test", "id": "fake_id"}`),
			success:        true,
			expectedResult: common.NewRequest().SetMethod("/test").SetID("fake_id"),
			expectedError:  nil,
		},
		{
			name:           "Number identifier",
			body:           []byte(`{"jsonrpc": "2.0", "method": "/test", "id": 4}`),
			success:        true,
			expectedResult: common.NewRequest().SetMethod("/test").SetID(4),
			expectedError:  nil,
		},
		{
			name:           "Null identifier",
			body:           []byte(`{"jsonrpc": "2.0", "method": "/test", "id": null}`),
			success:        true,
			expectedResult: common.NewRequest().SetMethod("/test").SetID(nil),
			expectedError:  nil,
		},
		{
			name:           "With param string",
			body:           []byte(`{"jsonrpc": "2.0", "method": "/test", "params": "test"}`),
			success:        true,
			expectedResult: common.NewRequest().SetMethod("/test").SetParams("test"),
			expectedError:  nil,
		},
		{
			name:           "With param number",
			body:           []byte(`{"jsonrpc": "2.0", "method": "/test", "params": 4}`),
			success:        true,
			expectedResult: common.NewRequest().SetMethod("/test").SetParams(float64(4)),
			expectedError:  nil,
		},
		{
			name:    "With param struct",
			body:    []byte(`{"jsonrpc": "2.0", "method": "/test", "params": {"foo": "bar", "baz": 4 }}`),
			success: true,
			expectedResult: common.NewRequest().SetMethod("/test").SetParams(map[string]interface{}{
				"foo": "bar",
				"baz": float64(4),
			}),
			expectedError: nil,
		},
		{
			name:    "Valid request : with param nested struct",
			body:    []byte(`{"jsonrpc": "2.0", "method": "/test", "params": {"foo": "bar", "baz": 4, "fizz": { "bool": true }}}`),
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
			body:    []byte(`{"jsonrpc": "2.0", "method": "/test", "id": "fake_id", "params": {"foo": "bar", "baz": 4, "fizz": { "bool": true }}}`),
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
			res, err := Request(tt.body)

			assert.Equal(t, tt.expectedResult, res)
			assert.Equal(t, tt.expectedError, err)
		})
	}
}
