package server

import (
	"context"
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
	req, err := parseRequest(r)
	if err != nil {
		common.NewResponse(nil).SetError(err).Send(w)
		return
	}

	common.NewResponse(req.ID).SetResult(req).Send(w)
	// Use dispatcher

	// Send response

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
