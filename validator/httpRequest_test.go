package validator

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHTTPRequest(t *testing.T) {
	testCases := []struct {
		name          string
		request       *http.Request
		success       bool
		expectedError error
	}{
		{
			name:          "Invalid method",
			request:       httptest.NewRequest("GET", "/", bytes.NewBuffer([]byte(`{"jsonrpc": "2.0", "id": 0, "method": "test"}`))),
			success:       false,
			expectedError: ErrInvalidHTTPMethod,
		},
		{
			name:          "Invalid path",
			request:       httptest.NewRequest("POST", "/unknown", bytes.NewBuffer([]byte(`{"jsonrpc": "2.0", "id": 0, "method": "test"}`))),
			success:       false,
			expectedError: ErrInvalidPathRequest,
		},
		{
			name:          "Valid request",
			request:       httptest.NewRequest("POST", "/", bytes.NewBuffer([]byte(`{"jsonrpc": "2.0", "id": 0, "method": "test"}`))),
			success:       true,
			expectedError: nil,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			err := HTTPRequest(tt.request)

			assert.Equal(t, tt.expectedError, err)
		})
	}
}
