package validator

import (
	"errors"
	"net/http"
)

var (
	ErrInvalidHTTPMethod  = errors.New("http method should be POST")
	ErrInvalidPathRequest = errors.New("http request should target /")
)

func HTTPRequest(r *http.Request) error {
	if r.Method != http.MethodPost {
		return ErrInvalidHTTPMethod
	}

	if r.URL.Path != "/" {
		return ErrInvalidPathRequest
	}

	return nil
}
