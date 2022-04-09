package validator

import (
	"errors"
	"reflect"

	"github.com/TomChv/jsonrpc2/common"
)

var (
	ErrInvalidJsonVersion    = errors.New("invalid JSON RPC version")
	ErrMissingMethod         = errors.New("no method provided")
	ErrInvalidIdentifierType = errors.New("http request invalid id type")
)

func JsonRPCRequest(req *common.Request) error {
	if req.ID != nil {
		switch reflect.TypeOf(req.ID).String() {
		case "string":
			break
		case "float64":
			// Verify if it's an integer or a float
			// nolint:forcetypeassert
			if req.ID != float64(int(req.ID.(float64))) {
				return ErrInvalidIdentifierType
			}
			// nolint:forcetypeassert
			req.SetID(int(req.ID.(float64)))
		default:
			req.SetID(nil)
			return ErrInvalidIdentifierType
		}
	}

	if req.JsonRpc != common.JSON_RPC_VERSION {
		return ErrInvalidJsonVersion
	}

	if req.Method == "" {
		return ErrMissingMethod
	}

	return nil
}
