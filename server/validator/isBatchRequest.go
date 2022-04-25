package validator

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
)

var (
	ErrMissingClosingBracket = errors.New("invalid batch request : missing closing bracket")
	ErrMissingOpeningBracket = errors.New("invalid batch request : missing opening bracket")
)

// IsBatchRequest return true if the request is wrapped with square brackets.
func IsBatchRequest(r *http.Request) (bool, error) {
	buf, _ := ioutil.ReadAll(r.Body)

	body := ioutil.NopCloser(bytes.NewReader(buf))
	data, err := ioutil.ReadAll(body)
	if err != nil {
		return false, err
	}

	// Reset body
	r.Body = ioutil.NopCloser(bytes.NewReader(buf))

	switch {
	case data[0] == '[' && data[len(data)-1] == ']':
		return true, nil
	case data[0] == '[':
		return false, ErrMissingClosingBracket
	case data[len(data)-1] == ']':
		return false, ErrMissingOpeningBracket
	default:
		return false, nil
	}
}
