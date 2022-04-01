package server

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"

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

// Implement HTTP interface to listen and response to incoming HTTP request
func (s *JsonRPC2) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := validator.HTTPRequest(r); err != nil {
		_ = common.NewResponse(nil).SetError(InvalidRequestError(err.Error())).Send(w)
		return
	}

	isBatch, err := validator.IsBatchRequest(r)
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

	ok, reqs, errs := parser.Batch(body)
	if !ok {
		_ = common.NewResponse(nil).SetError(InvalidRequestError("could not parse request")).Send(w)
		return
	}

	res := []*Response{}
	for _, err := range errs {
		res = append(res, common.NewResponse(nil).SetError(InvalidRequestError(err.Error())))
	}

	// Handle concurrency
	var (
		wg sync.WaitGroup
		l  sync.Mutex
	)

	for _, req := range reqs {
		req := req
		wg.Add(1)
		go func() {
			defer wg.Done()

			r := s.handle(req)

			l.Lock()
			defer l.Unlock()
			res = append(res, r)
		}()
	}

	wg.Wait()
	_ = s.sendBatch(res, w)
}

// Run start JSON RPC 2.0 server
func (s *JsonRPC2) Run(port string) error {
	addr := fmt.Sprintf(":%s", port)

	ctx, cancel := context.WithCancel(s.ctx)
	go func() {
		err := http.ListenAndServe(addr, s)
		if err != nil {
			log.Println(err)
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
