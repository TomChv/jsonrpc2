package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"

	"github.com/PtitLuca/go-dispatcher/dispatcher"
	"github.com/TomChv/jsonrpc2/common"
	"github.com/TomChv/jsonrpc2/server/parser"
	"github.com/TomChv/jsonrpc2/server/validator"
)

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
		_ = NewResponse(nil).SetError(InvalidRequestError(err)).Send(w)
		return
	}

	isBatch, err := validator.IsBatchRequest(r)
	if err != nil {
		_ = NewResponse(nil).SetError(ParsingError(err)).Send(w)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		_ = NewResponse(nil).SetError(ParsingError(err)).Send(w)
		return
	}

	if !isBatch {
		req, err := parser.Request(body)
		if err != nil {
			res := NewResponse(nil).SetError(InvalidRequestError(err))
			if req != nil && req.ID != nil {
				res.SetID(req.ID)
			}

			_ = res.Send(w)
			return
		}

		r := s.handle(req)
		if r.ID != nil {
			_ = r.Send(w)
		}
		return
	}

	reqs, err := parser.Batch(body)
	if err != nil {
		if errors.Is(err, parser.ErrEmptyBatch) {
			_ = NewResponse(nil).SetError(InvalidRequestError(err)).Send(w)
		} else {
			_ = NewResponse(nil).SetError(ParsingError(err)).Send(w)
		}
		return
	}

	batchRes := &Batch{}

	// Handle concurrency
	var wg sync.WaitGroup

	for _, rawReq := range reqs {
		wg.Add(1)

		go func(rawR []byte) {
			defer wg.Done()

			req := common.Request{}
			var r *Response

			err := json.Unmarshal(rawR, &req)
			if err != nil {
				r = NewResponse(nil).SetError(InvalidRequestError(err))
				batchRes.Append(r)
				return
			}

			if req.JsonRpc == "" {
				r = NewResponse(nil).SetError(InvalidRequestError(validator.ErrInvalidJsonVersion))
				batchRes.Append(r)
				return
			}

			if err := validator.JsonRPCRequest(&req); err != nil {
				r = NewResponse(nil).SetError(InvalidRequestError(err))
				if req.ID != nil {
					r.SetID(req.ID)
				}
			} else {
				r = s.handle(&req)
			}

			if r.ID == nil {
				return
			}

			batchRes.Append(r)
		}(rawReq)
	}

	wg.Wait()
	_ = batchRes.Send(w)
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
