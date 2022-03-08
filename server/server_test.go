package server

import (
	"context"
	"testing"

	"github.com/TomChv/jsonrpc2/common"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	s := New(context.TODO())
	assert.NotNilf(t, s, "Json RPC 2.0 server should not be nil")
}

type mockInvalidService struct{}

func (ms *mockInvalidService) MethodInvalidNoReturnType() {}

type mockInvalidService2 struct{}

func (ms *mockInvalidService2) MethodInvalidIncompleteReturnType() interface{} {
	return nil
}

type mockService struct{}

func (ms mockService) MethodEmptyArgs() (interface{}, error) {
	return nil, nil
}

func (ms mockService) MethodWithArg(str string) (string, error) {
	return str, nil
}

func (ms mockService) MethodWithArgs(str string, num int64) (map[string]interface{}, *RpcError) {
	return map[string]interface{}{
		"str": str,
		"num": num,
	}, nil
}

func (ms mockService) MethodWithComplexArgs(str []string, num int64, b bool, obj interface{}) (*Response, error) {
	return common.NewResponse("0").SetResult(struct {
		Str    []string
		Num    int64
		Bool   bool
		Object interface{}
	}{
		Str:    str,
		Num:    num,
		Bool:   b,
		Object: obj,
	}), nil
}

func TestJsonRPC2_Register(t *testing.T) {
	testCases := []struct {
		name          string
		serviceName   string
		service       interface{}
		success       bool
		expectedError error
	}{
		{
			name:          "Invalid service method : No return type",
			serviceName:   "mock",
			service:       &mockInvalidService{},
			success:       false,
			expectedError: ErrInvalidServiceProcedures,
		},
		{
			name:          "Invalid service method : Incomplete return type",
			serviceName:   "mock",
			service:       &mockInvalidService2{},
			success:       false,
			expectedError: ErrInvalidServiceProcedures,
		},
		{
			name:          "Valid service",
			serviceName:   "mock",
			service:       &mockService{},
			success:       true,
			expectedError: nil,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			s := New(context.TODO())
			err := s.Register(tt.serviceName, tt.service)

			assert.Equal(t, tt.expectedError, err)
		})
	}
}

func TestJsonRPC2_ServeHTTP(t *testing.T) {
	t.Skip("TODO")
}

func TestJsonRPC2_Run(t *testing.T) {
	//	ctx, cancel := context.WithCancel(context.TODO())
	//
	//	s := New(ctx)
	//	go func() {
	//		err := s.Run("8080")
	//		if err != nil {
	//			assert.Nil(t, err, "Run should not produce error")
	//		}
	//	}()
	//	cancel()
	t.Skip("don't know how to test it for now")
}
