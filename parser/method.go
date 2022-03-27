package parser

import (
	"errors"
	"strings"
)

var (
	ErrInvalidMethodFormat = errors.New("invalid method format")
)

// Procedure is a container around method easily parse which rpc must be called
type Procedure struct {
	Service string
	Method  string
}

// Method transforms a method string into an object Procedure
//   If there is no _, it returns the method with no service "" (e.g "sum")
//   If there is a _, it split the string and return the first part as a service
// and the second part as a method (e.g "eth_getBalance")
//   Otherwise, it returns an error
func Method(method string) (*Procedure, error) {
	toPascalCase := func(str string) string {
		if str == "" {
			return ""
		}
		return strings.ToUpper(string(str[0])) + str[1:]
	}

	switch strings.Count(method, "_") {
	case 0:
		return &Procedure{
			Method: toPascalCase(method),
		}, nil
	case 1:
		p := strings.Split(method, "_")
		return &Procedure{
			Service: p[0],
			Method:  toPascalCase(p[1]),
		}, nil
	default:
		return nil, ErrInvalidMethodFormat
	}
}
