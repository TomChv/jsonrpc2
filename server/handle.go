package server

import (
	"errors"

	"github.com/PtitLuca/go-dispatcher/dispatcher"
	"github.com/TomChv/jsonrpc2/common"
	"github.com/TomChv/jsonrpc2/parser"
)

var ErrNoFunctionErrorFound = errors.New("could not retrieve function error")

// handle json RPC 2 request :
//   - Retrieve procedure to call
//   - Convert arguments to their type
//   - Execute procedure
//   - Return response
func (s *JsonRPC2) handle(req *Request) *Response {
	p, err := parser.Method(req.Method)
	if err != nil {
		return common.NewResponse(req.ID).SetError(InvalidRequestError(err))
	}

	m, err := s.d.GetMethod(p.Service, p.Method)
	if err != nil {
		return common.NewResponse(req.ID).SetError(MethodNotFoundError(err))
	}

	args, err := parser.Arguments(m.GetArgsTypes()[1:], req.Params)
	if err != nil {
		return common.NewResponse(req.ID).SetError(InvalidParamsError(err))
	}

	// Run procedure
	ret, err := s.d.Run(p.Service, p.Method, args...)
	if err != nil {
		res := common.NewResponse(req.ID)

		switch {
		case errors.Is(err, dispatcher.ErrNonExistentMethod) ||
			errors.Is(err, dispatcher.ErrNonExistentService):
			res.SetError(MethodNotFoundError(err))
		case errors.Is(err, dispatcher.ErrInvalidArgumentType) ||
			errors.Is(err, dispatcher.ErrInvalidArgumentsCount):
			res.SetError(InvalidParamsError(err))
		default:
			res.SetError(InternalError(err))
		}

		return res
	}

	// Check for error
	if ret[1].Interface() != nil {
		err, ok := ret[1].Interface().(error)
		if !ok {
			return common.NewResponse(req.ID).SetError(InternalError(ErrNoFunctionErrorFound))
		}

		// Send error
		if err != nil {
			return common.NewResponse(req.ID).SetError(InternalError(err))
		}
	}

	// Send response
	return common.NewResponse(req.ID).SetResult(ret[0].Interface())
}
