package parser

import (
	"testing"

	"github.com/TomChv/jsonrpc2/common"
	"github.com/stretchr/testify/assert"
)

func TestBatch(t *testing.T) {
	testCases := []struct {
		name           string
		body           []byte
		success        bool
		expectedResult []*common.Request
		expectedError  error
	}{
		{
			name:           "Empty array",
			body:           []byte(`[]`),
			success:        true,
			expectedResult: []*common.Request{},
			expectedError:  nil,
		},
		{
			name:           "Simple request [Only Method]",
			body:           []byte(`[{"jsonrpc": "2.0", "method": "/test"}]`),
			success:        true,
			expectedResult: []*common.Request{common.NewRequest().SetMethod("/test")},
			expectedError:  nil,
		},
		{
			name:    "Simple request [Only Method, Invalid json rpc version]",
			body:    []byte(`[{"jsonrpc": "2.0", "method": "/test"},{"jsonrpc": "3.0","method": "/test"}]`),
			success: true,
			expectedResult: []*common.Request{
				common.NewRequest().SetMethod("/test"),
				{JsonRpc: "3.0", Method: "/test"},
			},
			expectedError: nil,
		},
		{
			name:    "Simple request [Missing method, Invalid json rpc version]",
			body:    []byte(`[{"jsonrpc": "2.0", "id": 0},{"jsonrpc": "3.0","method": "/test"}]`),
			success: true,
			expectedResult: []*common.Request{
				common.NewRequest().SetID(float64(0)),
				{JsonRpc: "3.0", Method: "/test"},
			},
			expectedError: nil,
		},
		{
			name:    "Simple request [Invalid json rpc version, Missing method]",
			body:    []byte(`[{"jsonrpc": "3.0","method": "/test"},{"jsonrpc": "2.0", "id": 0}]`),
			success: true,
			expectedResult: []*common.Request{
				{JsonRpc: "3.0", Method: "/test"},
				common.NewRequest().SetID(float64(0)),
			},
			expectedError: nil,
		},
		{
			name:    "Batch request [Only Method, Missing method, String identifier]",
			body:    []byte(`[{"jsonrpc": "2.0", "method": "/test"},{"jsonrpc": "2.0", "id": 0},{"jsonrpc": "2.0", "method": "/test", "id": "fake_id"}]`),
			success: true,
			expectedResult: []*common.Request{
				common.NewRequest().SetMethod("/test"),
				common.NewRequest().SetID(float64(0)),
				common.NewRequest().SetMethod("/test").SetID("fake_id"),
			},
			expectedError: nil,
		},
		{
			name:    "Batch request [Only Method, Missing method, String identifier, With param struct]",
			body:    []byte(`[{"jsonrpc": "2.0", "method": "/test"},{"jsonrpc": "2.0", "id": 0},{"jsonrpc": "2.0", "method": "/test", "id": "fake_id"},{"jsonrpc": "2.0", "method": "/test", "params": {"foo": "bar", "baz": 4 }}]`),
			success: true,
			expectedResult: []*common.Request{
				common.NewRequest().SetMethod("/test"),
				common.NewRequest().SetID(float64(0)),
				common.NewRequest().SetMethod("/test").SetID("fake_id"),
				common.NewRequest().SetMethod("/test").SetParams(map[string]interface{}{
					"foo": "bar",
					"baz": float64(4),
				}),
			},
			expectedError: nil,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			res, err := Batch(tt.body)

			assert.Equal(t, tt.expectedResult, res)
			assert.Equal(t, tt.expectedError, err)
		})
	}
}
