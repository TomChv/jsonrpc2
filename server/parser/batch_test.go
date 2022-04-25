package parser

import (
	"encoding/json"
	"testing"

	"github.com/TomChv/jsonrpc2/client"
	"github.com/stretchr/testify/assert"
)

func TestBatch(t *testing.T) {
	testCases := []struct {
		name           string
		body           []byte
		success        bool
		expectedResult []*client.Request
		expectedError  error
	}{
		{
			name:           "Empty array",
			body:           []byte(`[]`),
			success:        false,
			expectedResult: nil,
			expectedError:  ErrEmptyBatch,
		},
		{
			name:           "Simple request [Only Method]",
			body:           []byte(`[{"jsonrpc": "2.0", "method": "/test"}]`),
			success:        true,
			expectedResult: []*client.Request{client.NewRequest().SetMethod("/test")},
			expectedError:  nil,
		},
		{
			name:    "Simple request [Only Method, Invalid json rpc version]",
			body:    []byte(`[{"jsonrpc": "2.0", "method": "/test"},{"jsonrpc": "3.0","method": "/test"}]`),
			success: true,
			expectedResult: []*client.Request{
				client.NewRequest().SetMethod("/test"),
				{JsonRpc: "3.0", Method: "/test"},
			},
			expectedError: nil,
		},
		{
			name:    "Simple request [Missing method, Invalid json rpc version]",
			body:    []byte(`[{"jsonrpc": "2.0", "id": 0},{"jsonrpc": "3.0","method": "/test"}]`),
			success: true,
			expectedResult: []*client.Request{
				client.NewRequest().SetID(float64(0)),
				{JsonRpc: "3.0", Method: "/test"},
			},
			expectedError: nil,
		},
		{
			name:    "Simple request [Invalid json rpc version, Missing method]",
			body:    []byte(`[{"jsonrpc": "3.0","method": "/test"},{"jsonrpc": "2.0", "id": 0}]`),
			success: true,
			expectedResult: []*client.Request{
				{JsonRpc: "3.0", Method: "/test"},
				client.NewRequest().SetID(float64(0)),
			},
			expectedError: nil,
		},
		{
			name:    "Batch request [Only Method, Missing method, String identifier]",
			body:    []byte(`[{"jsonrpc": "2.0", "method": "/test"},{"jsonrpc": "2.0", "id": 0},{"jsonrpc": "2.0", "method": "/test", "id": "fake_id"}]`),
			success: true,
			expectedResult: []*client.Request{
				client.NewRequest().SetMethod("/test"),
				client.NewRequest().SetID(float64(0)),
				client.NewRequest().SetMethod("/test").SetID("fake_id"),
			},
			expectedError: nil,
		},
		{
			name:    "Batch request [Only Method, Missing method, String identifier, With param struct]",
			body:    []byte(`[{"jsonrpc": "2.0", "method": "/test"},{"jsonrpc": "2.0", "id": 0},{"jsonrpc": "2.0", "method": "/test", "id": "fake_id"},{"jsonrpc": "2.0", "method": "/test", "params": {"foo": "bar", "baz": 4 }}]`),
			success: true,
			expectedResult: []*client.Request{
				client.NewRequest().SetMethod("/test"),
				client.NewRequest().SetID(float64(0)),
				client.NewRequest().SetMethod("/test").SetID("fake_id"),
				client.NewRequest().SetMethod("/test").SetParams(map[string]interface{}{
					"foo": "bar",
					"baz": float64(4),
				}),
			},
			expectedError: nil,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			RawRes, err := Batch(tt.body)
			assert.Equal(t, tt.expectedError, err)

			if tt.expectedResult == nil {
				assert.Nil(t, RawRes)
				return
			}

			res := []*client.Request{}
			for _, rawR := range RawRes {
				var r *client.Request

				if err := json.Unmarshal(rawR, &r); err != nil {
					assert.Fail(t, err.Error())
				}
				res = append(res, r)
			}

			assert.Equal(t, tt.expectedResult, res)
		})
	}
}
