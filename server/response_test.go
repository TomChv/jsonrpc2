package server

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewResponse(t *testing.T) {
	testCases := []struct {
		name     string
		success  bool
		request  *Response
		expected string
	}{
		{
			name:     "Set string identifier",
			success:  true,
			request:  NewResponse("fake_id"),
			expected: `{"jsonrpc": "2.0", "id": "fake_id"}`,
		},
		{
			name:     "Set number identifier",
			success:  true,
			request:  NewResponse(4),
			expected: `{"jsonrpc": "2.0", "id": 4}`,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.request)

			assert.Nil(t, err, "marshal request should not fail")
			assert.JSONEq(t, string(data), tt.expected)
		})
	}
}

func TestResponse_SetID(t *testing.T) {
	testCases := []struct {
		name     string
		success  bool
		request  *Response
		expected string
	}{
		{
			name:     "Override identifier",
			success:  true,
			request:  NewResponse("fake_id").SetID(4),
			expected: `{"jsonrpc": "2.0", "id": 4}`,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.request)

			assert.Nil(t, err, "marshal request should not fail")
			assert.JSONEq(t, string(data), tt.expected)
		})
	}
}

func TestResponse_SetError(t *testing.T) {
	testCases := []struct {
		name     string
		success  bool
		request  *Response
		expected string
	}{
		{
			name:    "Set error : no data",
			success: true,
			request: NewResponse("fake_id").SetError(&RpcError{
				Code:    1000,
				Message: "error",
				Data:    nil,
			}),
			expected: `{"jsonrpc": "2.0", "id": "fake_id", "error": {"code": 1000, "message": "error"}}`,
		},
		{
			name:    "Set error : simple data",
			success: true,
			request: NewResponse("fake_id").SetError(&RpcError{
				Code:    1000,
				Message: "error",
				Data:    "additional context",
			}),
			expected: `{"jsonrpc": "2.0", "id": "fake_id", "error": {"code": 1000, "message": "error", "data": "additional context"}}`,
		},
		{
			name:    "Set error : complex data",
			success: true,
			request: NewResponse("fake_id").SetError(&RpcError{
				Code:    1000,
				Message: "error",
				Data: struct {
					Foo string `json:"foo"`
					Bar int    `json:"bar"`
				}{
					Foo: "bar",
					Bar: -1,
				},
			}),
			expected: `{"jsonrpc": "2.0", "id": "fake_id", "error": {"code": 1000, "message": "error", "data": {"foo": "bar", "bar": -1}}}`,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.request)

			assert.Nil(t, err, "marshal request should not fail")
			assert.JSONEq(t, tt.expected, string(data))
		})
	}
}

func TestResponse_SetResult(t *testing.T) {
	testCases := []struct {
		name     string
		success  bool
		request  *Response
		expected string
	}{
		{
			name:     "Set result : simple simple",
			success:  true,
			request:  NewResponse("fake_id").SetResult("data"),
			expected: `{"jsonrpc": "2.0", "id": "fake_id", "result": "data"}`,
		},
		{
			name:     "Set result : simple int",
			success:  true,
			request:  NewResponse("fake_id").SetResult(4),
			expected: `{"jsonrpc": "2.0", "id": "fake_id", "result": 4}`,
		},
		{
			name:    "Set result : struct",
			success: true,
			request: NewResponse("fake_id").SetResult(struct {
				Foo string `json:"foo"`
				Bar int    `json:"bar"`
			}{
				Foo: "foo",
				Bar: 12,
			}),
			expected: `{"jsonrpc": "2.0", "id": "fake_id", "result": {"foo": "foo", "bar": 12}}`,
		},
		{
			name:     "Set result : array",
			success:  true,
			request:  NewResponse("fake_id").SetResult([]string{"foo", "bar", "baz"}),
			expected: `{"jsonrpc": "2.0", "id": "fake_id", "result": ["foo", "bar", "baz"]}`,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.request)

			assert.Nil(t, err, "marshal request should not fail")
			assert.JSONEq(t, tt.expected, string(data))
		})
	}
}
