package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMethod(t *testing.T) {
	testCases := []struct {
		name           string
		method         string
		success        bool
		expectedResult *Procedure
		expectedError  error
	}{
		{
			name:           "Invalid method : 2 underscore",
			method:         "foo_bar_baz",
			success:        false,
			expectedResult: nil,
			expectedError:  ErrInvalidMethodFormat,
		},
		{
			name:           "Valid method : no underscore",
			method:         "foo",
			success:        true,
			expectedResult: &Procedure{Service: "", Method: "Foo"},
			expectedError:  nil,
		},
		{
			name:           "Valid method : no underscore and already PascalCase",
			method:         "Foo",
			success:        true,
			expectedResult: &Procedure{Service: "", Method: "Foo"},
			expectedError:  nil,
		},
		{
			name:           "Valid method : underscore",
			method:         "eth_getBalance",
			success:        true,
			expectedResult: &Procedure{Service: "eth", Method: "GetBalance"},
			expectedError:  nil,
		},
		{
			name:           "Valid method : underscore with already PascalCase",
			method:         "eth_GetBalance",
			success:        true,
			expectedResult: &Procedure{Service: "eth", Method: "GetBalance"},
			expectedError:  nil,
		},
		{
			name:           "Valid method : underscore with empty method",
			method:         "eth_",
			success:        true,
			expectedResult: &Procedure{Service: "eth", Method: ""},
			expectedError:  nil,
		},
		{
			name:           "Valid method : underscore with empty method",
			method:         "_getBalance",
			success:        true,
			expectedResult: &Procedure{Service: "", Method: "GetBalance"},
			expectedError:  nil,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			res, err := Method(tt.method)

			assert.Equal(t, tt.expectedResult, res)
			assert.Equal(t, tt.expectedError, err)
		})
	}
}
