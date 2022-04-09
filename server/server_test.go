package server

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/PtitLuca/go-dispatcher/dispatcher"
	"github.com/TomChv/jsonrpc2/common"
	"github.com/TomChv/jsonrpc2/validator"
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

func (ms mockService) MethodWithSleep(sec int64) (string, error) {
	time.Sleep(time.Second * time.Duration(sec))
	return "slept well", nil
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

func TestJsonRPC2_ServeHTTP_Single(t *testing.T) {
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
			req:              common.NewRequest().SetID("fake_id").SetMethod("mock_methodWithArgString").SetParams([]string{"bar"}),
			expectedResponse: common.NewResponse("fake_id").SetResult("bar"),
		},
		{
			name:             "call MethodWithArgNumber",
			success:          true,
			req:              common.NewRequest().SetID("fake_id").SetMethod("mock_methodWithArgNumber").SetParams([]int{4}),
			expectedResponse: common.NewResponse("fake_id").SetResult(float64(4)),
		},
		{
			name:             "call MethodWithArgFloat",
			success:          true,
			req:              common.NewRequest().SetID("fake_id").SetMethod("mock_methodWithArgFloat").SetParams([]float64{-2}),
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
			expectedResponse: common.NewResponse(nil).SetError(InvalidRequestError(validator.ErrMissingMethod.Error())),
		},
		{
			name:             "call with missing method and identifier",
			success:          true,
			req:              common.NewRequest().SetID("fake_id"),
			expectedResponse: common.NewResponse("fake_id").SetError(InvalidRequestError(validator.ErrMissingMethod.Error())),
		},
		{
			name:             "call with missing method and invalid identifier",
			success:          true,
			req:              common.NewRequest().SetID(false),
			expectedResponse: common.NewResponse(nil).SetError(InvalidRequestError(validator.ErrInvalidIdentifierType.Error())),
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

			var resData *Response
			err = json.Unmarshal(data, &resData)
			assert.Equal(t, nil, err)

			assert.Equal(t, tt.expectedResponse, resData)
		})
	}
}

func TestJsonRPC2_ServeHTTP_Batch(t *testing.T) {
	s := New(context.TODO())
	err := s.Register("mock", &mockService{})
	assert.Equal(t, nil, err)

	testCases := []struct {
		name              string
		success           bool
		reqs              []*Request
		expectedResponses []*Response
		timeout           time.Duration
	}{
		{
			name:              "Simple request [MethodEmptyArgs]",
			success:           true,
			reqs:              []*Request{common.NewRequest().SetID("fake_id").SetMethod("mock_methodEmptyArgs")},
			expectedResponses: []*Response{common.NewResponse("fake_id").SetResult("foo")},
		},
		{
			name:              "Simple request [Empty body]",
			success:           true,
			reqs:              []*Request{common.NewRequest()},
			expectedResponses: []*Response{common.NewResponse(nil).SetError(InvalidRequestError(validator.ErrMissingMethod.Error()))},
		},
		{
			name:    "Batch request [MethodEmptyArgs, Unknown Method]",
			success: true,
			reqs: []*Request{
				common.NewRequest().SetID("fake_id").SetMethod("mock_methodEmptyArgs"),
				common.NewRequest().SetID("fake_id").SetMethod("unknown"),
			},
			expectedResponses: []*Response{
				common.NewResponse("fake_id").SetError(MethodNotFoundError(dispatcher.ErrNonExistentService.Error())),
				common.NewResponse("fake_id").SetResult("foo"),
			},
		},
		{
			name:    "Batch request [MethodWithSleep, MethodWithSleep, MethodWithSleep] : test concurrency",
			success: true,
			reqs: []*Request{
				common.NewRequest().SetID("sleep_medium").SetMethod("mock_methodWithSleep").SetParams([]int64{3}),
				common.NewRequest().SetID("sleep_long").SetMethod("mock_methodWithSleep").SetParams([]int64{5}),
				common.NewRequest().SetID("sleep_fast").SetMethod("mock_methodWithSleep").SetParams([]int64{2}),
			},
			expectedResponses: []*Response{
				common.NewResponse("sleep_fast").SetResult("slept well"),
				common.NewResponse("sleep_medium").SetResult("slept well"),
				common.NewResponse("sleep_long").SetResult("slept well"),
			},
			timeout: time.Second * 7,
		},
		{
			name:    "Batch request [Unknown Method, MethodWithSleep, MethodWithSleep, MethodWithSleep, Empty body]",
			success: true,
			reqs: []*Request{
				common.NewRequest().SetID("fake_id").SetMethod("unknown"),
				common.NewRequest().SetID("sleep_medium").SetMethod("mock_methodWithSleep").SetParams([]int64{3}),
				common.NewRequest().SetID("sleep_long").SetMethod("mock_methodWithSleep").SetParams([]int64{5}),
				common.NewRequest().SetID("sleep_fast").SetMethod("mock_methodWithSleep").SetParams([]int64{1}),
				common.NewRequest(),
			},
			expectedResponses: []*Response{
				common.NewResponse(nil).SetError(InvalidRequestError(validator.ErrMissingMethod.Error())),
				common.NewResponse("fake_id").SetError(MethodNotFoundError(dispatcher.ErrNonExistentService.Error())),
				common.NewResponse("sleep_fast").SetResult("slept well"),
				common.NewResponse("sleep_medium").SetResult("slept well"),
				common.NewResponse("sleep_long").SetResult("slept well"),
			},
			timeout: time.Second * 7,
		},
		{
			name:    "Simple request [Only Method, Invalid json rpc version]",
			success: true,
			reqs: []*Request{
				common.NewRequest().SetID("fake_id").SetMethod("mock_methodEmptyArgs"),
				{JsonRpc: "3.0", Method: "/test"},
			},
			expectedResponses: []*Response{
				common.NewResponse(nil).SetError(InvalidRequestError(validator.ErrInvalidJsonVersion.Error())),
				common.NewResponse("fake_id").SetResult("foo"),
			},
		},
		{
			name:    "Simple request [Invalid json rpc version, Missing method]",
			success: true,
			reqs: []*Request{
				{JsonRpc: "3.0", Method: "/test"},
				common.NewRequest().SetID(0),
			},
			expectedResponses: []*Response{
				common.NewResponse(float64(0)).SetError(InvalidRequestError(validator.ErrMissingMethod.Error())),
				common.NewResponse(nil).SetError(InvalidRequestError(validator.ErrInvalidJsonVersion.Error())),
			},
		},
		{
			name:    "Simple request [MethodWithArgString, Missing method, String identifier, MethodWithArgs]",
			success: true,
			reqs: []*Request{
				common.NewRequest().SetID(4).SetMethod("mock_methodWithArgString").SetParams([]string{"foo"}),
				common.NewRequest().SetID(0),
				common.NewRequest().SetID("fake_id_number").SetMethod("mock_methodWithArgNumber").SetParams([]int64{687}),
				common.NewRequest().SetID("fake_id_args").SetMethod("mock_methodWithArgs").SetParams([]interface{}{"foo", -1}),
			},
			expectedResponses: []*Response{
				common.NewResponse(float64(0)).SetError(InvalidRequestError(validator.ErrMissingMethod.Error())),
				common.NewResponse("fake_id_args").SetResult(map[string]interface{}{
					"str": "foo",
					"num": float64(-1),
				}),
				common.NewResponse(float64(4)).SetResult("foo"),
				common.NewResponse("fake_id_number").SetResult(float64(687)),
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.reqs)
			assert.Equal(t, nil, err)

			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
			w := httptest.NewRecorder()

			start := time.Now()
			s.ServeHTTP(w, req)

			res := w.Result()
			defer res.Body.Close()

			if tt.timeout != 0 && time.Since(start) > tt.timeout {
				assert.Fail(t, "test reached timeout")
			}

			data, err := ioutil.ReadAll(res.Body)
			assert.Equal(t, nil, err)

			var resData []*Response
			err = json.Unmarshal(data, &resData)
			assert.Equal(t, nil, err)

			assert.ElementsMatch(t, tt.expectedResponses, resData)
			//assert.Equal(t, , )
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
