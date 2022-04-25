package validator

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsBatchRequest(t *testing.T) {
	testCases := []struct {
		name           string
		success        bool
		request        *http.Request
		expectedResult bool
		expectedError  error
	}{
		{
			name:           "Single call request",
			success:        true,
			request:        httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(`{"jsonrpc": "3.0", "id": 0, "method": "test"}`))),
			expectedResult: false,
			expectedError:  nil,
		},
		{
			name:           "Batch call request",
			success:        true,
			request:        httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(`[{"jsonrpc": "3.0", "id": 0, "method": "test"},{"jsonrpc": "3.0", "id": 0, "method": "test"}]`))),
			expectedResult: true,
			expectedError:  nil,
		},
		{
			name:           "Missing opening square bracket",
			success:        true,
			request:        httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(`{"jsonrpc": "3.0", "id": 0, "method": "test"},{"jsonrpc": "3.0", "id": 0, "method": "test"}]`))),
			expectedResult: false,
			expectedError:  ErrMissingOpeningBracket,
		},
		{
			name:           "Missing closing square bracket",
			success:        true,
			request:        httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(`[{"jsonrpc": "3.0", "id": 0, "method": "test"},{"jsonrpc": "3.0", "id": 0, "method": "test"}`))),
			expectedResult: false,
			expectedError:  ErrMissingClosingBracket,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			isBatch, err := IsBatchRequest(tt.request)

			assert.Equal(t, tt.expectedResult, isBatch)
			assert.Equal(t, tt.expectedError, err)
		})
	}
}
