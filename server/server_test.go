package server

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/PtitLuca/go-dispatcher/dispatcher"
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
	return "foo", nil
}

func (ms mockService) MethodWithArgString(str string) (string, error) {
	return str, nil
}

func (ms mockService) MethodWithArgNumber(num int) (int, error) {
	return num, nil
}

func (ms mockService) MethodWithArgFloat(float float64) (float64, error) {
	return float, nil
}

func (ms mockService) MethodWithArgs(str string, num int64) (map[string]interface{}, error) {
	return map[string]interface{}{
		"str": str,
		"num": num,
	}, nil
}

type FakeStruct struct {
	Id     int
	Field1 bool
	Field2 string
}

func (ms mockService) MethodWithComplexArgs(str []string, num int8, b bool, obj FakeStruct) (map[string]interface{}, error) {
	return map[string]interface{}{
		"str":    str,
		"num":    num,
		"bool":   b,
		"object": obj,
	}, nil
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
	s := New(context.TODO())
	err := s.Register("mock", &mockService{})
	assert.Equal(t, nil, err)

	testCases := []struct {
		name             string
		success          bool
		req              *Request
		expectedResponse *Response
	}{
		{
			name:             "call MethodEmptyArgs",
			success:          true,
			req:              common.NewRequest().SetID("fake_id").SetMethod("mock_methodEmptyArgs"),
			expectedResponse: common.NewResponse("fake_id").SetResult("foo"),
		},
		{
			name:             "call MethodWithArgString",
			success:          true,
			req:              common.NewRequest().SetID("fake_id").SetMethod("mock_methodWithArgString").SetParams("bar"),
			expectedResponse: common.NewResponse("fake_id").SetResult("bar"),
		},
		{
			name:             "call MethodWithArgNumber",
			success:          true,
			req:              common.NewRequest().SetID("fake_id").SetMethod("mock_methodWithArgNumber").SetParams(4),
			expectedResponse: common.NewResponse("fake_id").SetResult(float64(4)),
		},
		{
			name:             "call MethodWithArgFloat",
			success:          true,
			req:              common.NewRequest().SetID("fake_id").SetMethod("mock_methodWithArgFloat").SetParams(-2),
			expectedResponse: common.NewResponse("fake_id").SetResult(float64(-2)),
		},
		{
			name:    "call MethodWithArgs",
			success: true,
			req:     common.NewRequest().SetID("fake_id").SetMethod("mock_methodWithArgs").SetParams([]interface{}{"foo", 25}),
			expectedResponse: common.NewResponse("fake_id").SetResult(map[string]interface{}{
				"str": "foo",
				"num": float64(25),
			}),
		},
		{
			name:    "call MethodWithComplexArgs",
			success: true,
			req: common.NewRequest().SetID("fake_id").SetMethod("mock_methodWithComplexArgs").SetParams([]interface{}{
				[]string{"foo", "bar"},
				25,
				false,
				FakeStruct{0, true, "fakeStruct"},
			}),
			expectedResponse: common.NewResponse("fake_id").SetResult(map[string]interface{}{
				"str":  []interface{}{"foo", "bar"},
				"num":  float64(25),
				"bool": false,
				"object": map[string]interface{}{
					"Id":     float64(0),
					"Field1": true,
					"Field2": "fakeStruct",
				},
			}),
		},
		{
			name:             "call MethodEmptyArgs with int identifier",
			success:          true,
			req:              common.NewRequest().SetID(-1).SetMethod("mock_methodEmptyArgs"),
			expectedResponse: common.NewResponse(float64(-1)).SetResult("foo"),
		},
		{
			name:             "call MethodEmptyArgs with no identifier",
			success:          true,
			req:              common.NewRequest().SetMethod("mock_methodEmptyArgs"),
			expectedResponse: common.NewResponse(nil).SetResult("foo"),
		},
		{
			name:             "call unknown method",
			success:          false,
			req:              common.NewRequest().SetID("fake_id").SetMethod("unknown"),
			expectedResponse: common.NewResponse("fake_id").SetError(MethodNotFoundError(dispatcher.ErrNonExistentService.Error())),
		},
		{
			name:             "call with empty body",
			success:          false,
			req:              common.NewRequest(),
			expectedResponse: common.NewResponse(nil).SetError(InvalidRequestError(ErrNoMethodProvided.Error())),
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			body, err := tt.req.Bytes()
			assert.Equal(t, nil, err)

			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
			w := httptest.NewRecorder()

			s.ServeHTTP(w, req)

			res := w.Result()
			defer res.Body.Close()
			data, err := ioutil.ReadAll(res.Body)
			assert.Equal(t, nil, err)

			var resData Response
			err = json.Unmarshal(data, &resData)
			assert.Equal(t, nil, err)

			if resData.Result != nil {
				assert.Equal(t, *tt.expectedResponse, resData)
			} else {
				assert.Equal(t, tt.expectedResponse.Error, resData.Error)
			}
		})
	}
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
