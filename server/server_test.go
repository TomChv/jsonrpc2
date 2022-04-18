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
			name:             "call MethodEmptyArgs with notification",
			success:          true,
			req:              common.NewRequest().SetMethod("mock_methodEmptyArgs"),
			expectedResponse: nil,
		},
		{
			name:             "call MethodWithArgs",
			success:          true,
			req:              common.NewRequest().SetMethod("mock_methodWithArgs").SetParams([]interface{}{"foo", 25}),
			expectedResponse: nil,
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

			if tt.expectedResponse == nil {
				assert.Equal(t, int64(-1), res.ContentLength)
				return
			}

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
			expectedResponses: nil,
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
			},
			expectedResponses: []*Response{
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
			},
			expectedResponses: []*Response{
				common.NewResponse("fake_id").SetResult("foo"),
			},
		},
		{
			name:    "Simple request [Invalid json rpc version, Missing method]",
			success: true,
			reqs: []*Request{
				common.NewRequest().SetID(0),
			},
			expectedResponses: []*Response{
				common.NewResponse(float64(0)).SetError(InvalidRequestError(validator.ErrMissingMethod.Error())),
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
			if tt.expectedResponses == nil {
				assert.Equal(t, int64(-1), res.ContentLength)
				return
			}

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
		})
	}
}

type officialExample struct{}
type officialNotificationExample struct{}
type officialGetExample struct{}

func (o officialExample) PositionalSubtract(a, b int) (interface{}, error) {
	return a - b, nil
}

func (o officialExample) NamedSubtract(arg struct {
	Minuend    int `json:"minuend"`
	Subtrahend int `json:"subtrahend"`
}) (interface{}, error) {
	return arg.Minuend - arg.Subtrahend, nil
}

func (o officialExample) Update(args []int) (interface{}, error) {
	return args, nil
}

func (o officialExample) Sum(args []int) (int, error) {
	res := 0
	for _, arg := range args {
		res += arg
	}
	return res, nil
}

func (o officialNotificationExample) Hello(arg int) (int, error) {
	return arg, nil
}

func (o officialGetExample) Data() ([]interface{}, error) {
	return []interface{}{"hello", 5}, nil
}

func TestJsonRPC2_ServeHTTP_Official_Example(t *testing.T) {
	s := New(context.TODO())
	err := s.Register("", &officialExample{})
	assert.Equal(t, nil, err)
	err = s.Register("notify", &officialNotificationExample{})
	assert.Equal(t, nil, err)
	err = s.Register("get", &officialGetExample{})
	assert.Equal(t, nil, err)

	testCases := []struct {
		name             string
		success          bool
		batch            bool
		req              []byte
		expectedResponse []byte
	}{
		{
			name:             "RPC call with positional parameters - subtract",
			success:          true,
			req:              []byte(`{"jsonrpc": "2.0", "method": "positionalSubtract", "params": [42, 23], "id": 1}`),
			expectedResponse: []byte(`{"jsonrpc":"2.0","result":19,"id":1}`),
		},
		{
			name:             "RPC call with positional parameters - reverse subtract",
			success:          true,
			req:              []byte(`{"jsonrpc": "2.0", "method": "positionalSubtract", "params": [23, 42], "id": 2}`),
			expectedResponse: []byte(`{"jsonrpc":"2.0","result":-19,"id":2}`),
		},
		{
			name:             "RPC call with named parameters - subtract",
			success:          true,
			req:              []byte(`{"jsonrpc": "2.0", "method": "namedSubtract", "params": {"subtrahend": 23, "minuend": 42}, "id": 3}`),
			expectedResponse: []byte(`{"jsonrpc":"2.0","result":19,"id":3}`),
		},
		{
			name:             "RPC call with named parameters - reverse subtract",
			success:          true,
			req:              []byte(`{"jsonrpc": "2.0", "method": "namedSubtract", "params": {"minuend": 42, "subtrahend": 23}, "id": 4}`),
			expectedResponse: []byte(`{"jsonrpc":"2.0","result":19,"id":4}`),
		},
		{
			name:             "RPC call notification - update",
			success:          true,
			req:              []byte(`{"jsonrpc": "2.0", "method": "update", "params": [1,2,3,4,5]}`),
			expectedResponse: nil,
		},
		{
			name:             "RPC call notification  of non-existent method - foobar",
			success:          true,
			req:              []byte(`{"jsonrpc": "2.0", "method": "foobar"}`),
			expectedResponse: nil,
		},
		{
			name:             "RPC call of non-existent method - foobar",
			success:          false,
			req:              []byte(`{"jsonrpc": "2.0", "method": "foobar", "id": "1"}`),
			expectedResponse: []byte(`{"jsonrpc": "2.0", "error": {"code": -32601, "message": "Method not found"}, "id": "1"}`),
		},
		{
			name:             "RPC call of invalid JSON - foobar",
			success:          false,
			req:              []byte(`{"jsonrpc": "2.0", "method": "foobar, "params": "bar", "baz]`),
			expectedResponse: []byte(`{"jsonrpc": "2.0", "error": {"code": -32700, "message": "Parse error"}, "id": null}`),
		},
		{
			name:             "RPC call with invalid Request Object",
			success:          false,
			req:              []byte(`{"jsonrpc": "2.0", "method": 1, "params": "bar"}`),
			expectedResponse: []byte(`{"jsonrpc": "2.0", "error": {"code": -32600, "message": "Invalid Request"}, "id": null}`),
		},
		{
			name:             "RPC call Batch - invalid JSON",
			success:          false,
			req:              []byte(`[{"jsonrpc": "2.0", "method": "sum", "params": [1,2,4], "id": "1"},{"jsonrpc": "2.0", "method"]`),
			expectedResponse: []byte(`{"jsonrpc": "2.0", "error": {"code": -32700, "message": "Parse error"}, "id": null}`),
		},
		{
			name:             "RPC call with an empty Array",
			success:          false,
			req:              []byte(`[]`),
			expectedResponse: []byte(`{"jsonrpc": "2.0", "error": {"code": -32600, "message": "Invalid Request"}, "id": null}`),
		},
		{
			name:             "RPC call with an invalid Batch - one element",
			success:          false,
			req:              []byte(`[1]`),
			expectedResponse: []byte(`[{"jsonrpc": "2.0", "error": {"code": -32600, "message": "Invalid Request"}, "id": null}]`),
			batch:            true,
		},
		{
			name:             "RPC call with an invalid Batch - multi element",
			success:          false,
			req:              []byte(`[1,2,3]`),
			expectedResponse: []byte(`[{"jsonrpc": "2.0", "error": {"code": -32600, "message": "Invalid Request"}, "id": null},{"jsonrpc": "2.0", "error": {"code": -32600, "message": "Invalid Request"}, "id": null},{"jsonrpc": "2.0", "error": {"code": -32600, "message": "Invalid Request"}, "id": null}]`),
			batch:            true,
		},
		{
			name:             "RPC call Batch",
			success:          false,
			req:              []byte(`[{"jsonrpc": "2.0", "method": "sum", "params": [1,2,4], "id": "1"},{"jsonrpc": "2.0", "method": "notify_hello", "params": [7]},{"jsonrpc": "2.0", "method": "positionalSubtract", "params": [42,23], "id": "2"},{"foo": "boo"},{"jsonrpc": "2.0", "method": "foo.get", "params": {"name": "myself"}, "id": "5"},{"jsonrpc": "2.0", "method": "get_data", "id": "9"}]`),
			expectedResponse: []byte(`[{"jsonrpc": "2.0", "result": 7, "id": "1"},{"jsonrpc": "2.0", "result": 19, "id": "2"},{"jsonrpc": "2.0", "error": {"code": -32600, "message": "Invalid Request"}, "id": null},{"jsonrpc": "2.0", "error": {"code": -32601, "message": "Method not found"}, "id": "5"},{"jsonrpc": "2.0", "result": ["hello", 5], "id": "9"}]`),
			batch:            true,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(tt.req))
			w := httptest.NewRecorder()

			s.ServeHTTP(w, req)

			res := w.Result()

			defer res.Body.Close()

			data, err := ioutil.ReadAll(res.Body)
			if tt.expectedResponse == nil {
				assert.Empty(t, data)
				return
			}

			assert.Equal(t, nil, err)
			if tt.success {
				assert.Equal(t, string(tt.expectedResponse), string(data))
				return
			}

			if !tt.batch {
				var r common.Response
				var expect common.Response

				_ = json.Unmarshal(tt.expectedResponse, &expect)
				_ = json.Unmarshal(data, &r)

				assert.Equal(t, expect.ID, r.ID)
				assert.Equal(t, expect.Error.Code, r.Error.Code)
				assert.Equal(t, expect.Error.Message, r.Error.Message)
			} else {
				var res []common.Response
				var expects []common.Response

				_ = json.Unmarshal(tt.expectedResponse, &expects)
				_ = json.Unmarshal(data, &res)

				// Ignore data
				for _, r := range res {
					if r.Error != nil {
						r.Error.Data = nil
					}
				}

				assert.ElementsMatch(t, expects, res)
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
