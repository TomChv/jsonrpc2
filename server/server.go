package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/PtitLuca/go-dispatcher/dispatcher"
	"github.com/TomChv/jsonrpc2/common"
)

type Response = common.Response
type Request = common.Request
type RpcError = common.RpcError

// JsonRPC2 is a simple HTTP server that follow JSON RPC 2.0 specification
// See https://www.jsonrpc.org/specification for more information
type JsonRPC2 struct {
	ctx context.Context
	d   *dispatcher.Dispatcher
}

// New create a JSON RPC 2.0 server
func New(ctx context.Context) *JsonRPC2 {
	return &JsonRPC2{
		ctx: ctx,
		d:   dispatcher.New(),
	}
}

// Register a new RPC
//
// Not matter what your service's procedures takes as parameters they
// must return (*Response, *RpcError)
func (s *JsonRPC2) Register(namespace string, service interface{}) error {
	if !validateService(service) {
		return ErrInvalidServiceProcedures
	}

	if err := s.d.Register(namespace, service); err != nil {
		return err
	}
	return nil
}

// Implement HTTP interface to listen and response to incoming HTTP request
func (s *JsonRPC2) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Parse request
	req, rpcErr := parseRequest(r)
	if rpcErr != nil {
		_ = common.NewResponse(nil).SetError(rpcErr).Send(w)
		return
	}

	// Use dispatcher
	p, rpcErr := parseMethod(req.Method)
	if rpcErr != nil {
		_ = common.NewResponse(req.ID).SetError(rpcErr).Send(w)
		return
	}

	m, err := s.d.GetMethod(p.Service, p.Method)
	if err != nil {
		_ = common.NewResponse(req.ID).SetError(MethodNotFoundError(err.Error())).Send(w)
		return
	}

	args, rpcErr := parseParams(m.GetArgsTypes()[1:], req.Params)
	if rpcErr != nil {
		_ = common.NewResponse(req.ID).SetError(InvalidParamsError(rpcErr.Error())).Send(w)
		return
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

		_ = res.Send(w)
		return
	}

	// Check for error
	if ret[1].Interface() != nil {
		errCall, ok := ret[1].Interface().(error)
		if !ok {
			_ = common.NewResponse(req.ID).SetError(InternalError("could not retrieve function error")).Send(w)
			return
		}

		// Send error
		if errCall != nil {
			_ = common.NewResponse(req.ID).SetError(InternalError(errCall)).Send(w)
			return
		}
	}

	// Send response
	_ = common.NewResponse(req.ID).SetResult(ret[0].Interface()).Send(w)
}

// Run start JSON RPC 2.0 server
func (s *JsonRPC2) Run(port string) error {
	addr := fmt.Sprintf(":%s", port)

	ctx, cancel := context.WithCancel(s.ctx)
	go func() {
		err := http.ListenAndServe(addr, s)
		if err != nil {
			cancel()
		}
	}()

	log.Println(fmt.Sprintf("JSON RPC 2.0 server listening on http://0.0.0.0:%s", addr))

	select {
	case <-s.ctx.Done():
		cancel()
	case <-ctx.Done():
		return nil
	}
	return nil
}
