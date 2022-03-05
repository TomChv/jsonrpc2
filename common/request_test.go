package common

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRequest(t *testing.T) {
	req := NewRequest()
	data, err := json.Marshal(req)

	assert.Nil(t, err, "marshal request should not fail")
	assert.JSONEq(t, string(data), `{"jsonrpc": "2.0", "method": ""}`)
}

func TestRequest_SetID(t *testing.T) {
	testCases := []struct {
		name     string
		success  bool
		request  *Request
		expected string
	}{
		{
			name:     "Set string identifier",
			success:  true,
			request:  NewRequest().SetID("fake_id"),
			expected: `{"jsonrpc": "2.0", "method": "", "id": "fake_id"}`,
		},
		{
			name:     "Set number identifier",
			success:  true,
			request:  NewRequest().SetID(4),
			expected: `{"jsonrpc": "2.0", "method": "", "id": 4}`,
		},
		{
			name:     "Override identifier",
			success:  true,
			request:  NewRequest().SetID(4).SetID("fake_id"),
			expected: `{"jsonrpc": "2.0", "method": "", "id": "fake_id"}`,
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

func TestRequest_SetMethod(t *testing.T) {
	testCases := []struct {
		name     string
		success  bool
		request  *Request
		expected string
	}{
		{
			name:     "Set method",
			success:  true,
			request:  NewRequest().SetMethod("test"),
			expected: `{"jsonrpc": "2.0", "method": "test"}`,
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

func TestRequest_SetParams(t *testing.T) {
	testCases := []struct {
		name     string
		success  bool
		request  *Request
		expected string
	}{
		{
			name:     "Set params : string",
			success:  true,
			request:  NewRequest().SetParams("test"),
			expected: `{"jsonrpc": "2.0", "method": "", "params": "test"}`,
		},
		{
			name:     "Set params : int",
			success:  true,
			request:  NewRequest().SetParams(5),
			expected: `{"jsonrpc": "2.0", "method": "", "params": 5}`,
		},
		{
			name:     "Set params : bool",
			success:  true,
			request:  NewRequest().SetParams(false),
			expected: `{"jsonrpc": "2.0", "method": "", "params": false}`,
		},
		{
			name:    "Set params : struct",
			success: true,
			request: NewRequest().SetParams(struct {
				Foo string `json:"foo"`
				Bar int    `json:"bar"`
			}{
				Foo: "foo",
				Bar: 5,
			}),
			expected: `{"jsonrpc": "2.0", "method": "", "params": {"foo": "foo", "bar": 5}}`,
		},
		{
			name:    "Set params : nested struct ",
			success: true,
			request: NewRequest().SetParams(struct {
				Foo string `json:"foo"`
				Bar []int  `json:"bar"`
				Baz struct {
					Bool bool     `json:"bool"`
					Dog  []string `json:"dog"`
				} `json:"baz"`
			}{
				Foo: "foo",
				Bar: []int{5, 6},
				Baz: struct {
					Bool bool     `json:"bool"`
					Dog  []string `json:"dog"`
				}{
					Bool: false,
					Dog:  []string{"baz", "baz"},
				},
			}),
			expected: `{"jsonrpc": "2.0", "method": "", "params": {"bar": [5, 6], "baz": {"bool": false, "dog": ["baz", "baz"]}, "foo": "foo"}}`,
		},
		{
			name:     "Set params : array string",
			success:  true,
			request:  NewRequest().SetParams([2]string{"foo", "bar"}),
			expected: `{"jsonrpc": "2.0", "method": "", "params": ["foo","bar"]}`,
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

func TestRequest_SetAllFields(t *testing.T) {
	testCases := []struct {
		name     string
		success  bool
		request  *Request
		expected string
	}{
		{
			name:     "Set simple request",
			success:  true,
			request:  NewRequest().SetID("fake_id").SetMethod("test").SetParams("test"),
			expected: `{"jsonrpc": "2.0", "method": "test", "id": "fake_id", "params": "test"}`,
		},
		{
			name:    "Set complex request",
			success: true,
			request: NewRequest().SetParams(struct {
				Foo string `json:"foo"`
				Bar []int  `json:"bar"`
				Baz struct {
					Bool bool     `json:"bool"`
					Dog  []string `json:"dog"`
				} `json:"baz"`
			}{
				Foo: "foo",
				Bar: []int{5, 6},
				Baz: struct {
					Bool bool     `json:"bool"`
					Dog  []string `json:"dog"`
				}{
					Bool: false,
					Dog:  []string{"baz", "baz"},
				},
			}).SetID(4).SetMethod("complex"),
			expected: `{"jsonrpc": "2.0", "method": "complex", "id": 4, "params": {"bar": [5, 6], "baz": {"bool": false, "dog": ["baz", "baz"]}, "foo": "foo"}}`,
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
