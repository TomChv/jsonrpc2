package server

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/PtitLuca/go-dispatcher/dispatcher"
	"github.com/TomChv/jsonrpc2/common"
	"github.com/TomChv/jsonrpc2/parser"
	"github.com/TomChv/jsonrpc2/validator"
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

// isBatchRequest return true if the request is wrapped with square brackets.
func isBatchRequest(r *http.Request) (bool, error) {
	buf, _ := ioutil.ReadAll(r.Body)

	body := ioutil.NopCloser(bytes.NewReader(buf))
	data, err := ioutil.ReadAll(body)
	if err != nil {
		return false, err
	}

	// Reset body
	r.Body = ioutil.NopCloser(bytes.NewReader(buf))
	return data[0] == '[' && data[len(data)-1] == ']', nil
}

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

// Implement HTTP interface to listen and response to incoming HTTP request
func (s *JsonRPC2) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := validator.HTTPRequest(r); err != nil {
		_ = common.NewResponse(nil).SetError(InvalidRequestError(err.Error())).Send(w)
		return
	}

	isBatch, err := isBatchRequest(r)
	if err != nil {
		_ = common.NewResponse(nil).SetError(ParsingError(err.Error())).Send(w)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		_ = common.NewResponse(nil).SetError(ParsingError(err.Error())).Send(w)
		return
	}

	if !isBatch {
		req, err := parser.Request(body)
		if err != nil {
			_ = common.NewResponse(nil).SetError(InvalidRequestError(err.Error())).Send(w)
			return
		}

		_ = s.handle(req).Send(w)
		return
	}
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
