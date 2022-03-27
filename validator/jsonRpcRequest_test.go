package validator

import (
	"testing"

	"github.com/TomChv/jsonrpc2/common"
	"github.com/stretchr/testify/assert"
)

func TestJsonRPCRequest(t *testing.T) {
	testCases := []struct {
		name          string
		request       *common.Request
		success       bool
		expectedError error
	}{
		{
			name:          "Invalid JSON RPC version",
			request:       &common.Request{},
			success:       false,
			expectedError: ErrInvalidJsonVersion,
		},
		{
			name:          "Missing method",
			request:       common.NewRequest(),
			success:       false,
			expectedError: ErrMissingMethod,
		},
		{
			name:          "Invalid identifier type - Boolean",
			request:       common.NewRequest().SetMethod("test").SetID(true),
			success:       false,
			expectedError: ErrInvalidIdentifierType,
		},
		{
			name:          "Invalid identifier type - Struct",
			request:       common.NewRequest().SetMethod("test").SetID(map[string]interface{}{"foo": "bar"}),
			success:       false,
			expectedError: ErrInvalidIdentifierType,
		},
		{
			name:          "Only method",
			request:       common.NewRequest().SetMethod("test"),
			success:       true,
			expectedError: nil,
		},
		{
			name:          "String identifier",
			request:       common.NewRequest().SetMethod("test").SetID("foo"),
			success:       true,
			expectedError: nil,
		},
		{
			name:          "Number identifier",
			request:       common.NewRequest().SetMethod("test").SetID(float64(4)),
			success:       true,
			expectedError: nil,
		},
		{
			name:          "Null identifier",
			request:       common.NewRequest().SetMethod("test").SetID(nil),
			success:       true,
			expectedError: nil,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			err := JsonRPCRequest(tt.request)

			assert.Equal(t, tt.expectedError, err)
		})
	}
}
