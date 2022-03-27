package server

import (
	"errors"

	"github.com/PtitLuca/go-dispatcher/dispatcher"
	"github.com/TomChv/jsonrpc2/common"
	"github.com/TomChv/jsonrpc2/parser"
)

// handle json RPC 2 request :
//   - Retrieve procedure to call
//   - Convert arguments to their type
//   - Execute procedure
//   - Return response
func (s *JsonRPC2) handle(req *Request) *Response {
	p, err := parser.Method(req.Method)
	if err != nil {
		return common.NewResponse(req.ID).SetError(InvalidRequestError(err.Error()))
	}

	m, err := s.d.GetMethod(p.Service, p.Method)
	if err != nil {
		return common.NewResponse(req.ID).SetError(MethodNotFoundError(err.Error()))
	}

	args, err := parser.Arguments(m.GetArgsTypes()[1:], req.Params)
	if err != nil {
		return common.NewResponse(req.ID).SetError(InvalidParamsError(err.Error()))
	}

	// Run procedure
	ret, errCall := s.d.Run(p.Service, p.Method, args...)
	if errCall != nil {
		res := common.NewResponse(req.ID)

		switch {
		case errors.Is(errCall, dispatcher.ErrNonExistentMethod) ||
			errors.Is(errCall, dispatcher.ErrNonExistentService):
			res.SetError(MethodNotFoundError(errCall.Error()))
		case errors.Is(errCall, dispatcher.ErrInvalidArgumentType) ||
			errors.Is(errCall, dispatcher.ErrInvalidArgumentsCount):
			res.SetError(InvalidParamsError(errCall.Error()))
		default:
			res.SetError(InternalError(errCall.Error()))
		}

		return res
	}

	// Check for error
	if ret[1].Interface() != nil {
		errCall, ok := ret[1].Interface().(error)
		if !ok {
			return common.NewResponse(req.ID).SetError(InternalError("could not retrieve function error"))
		}

		// Send error
		if errCall != nil {
			return common.NewResponse(req.ID).SetError(InternalError(errCall))
		}
	}

	// Send response
	return common.NewResponse(req.ID).SetResult(ret[0].Interface())
}
