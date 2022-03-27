package parser

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestArguments(t *testing.T) {
	type FakeStruct struct {
		Id     int
		Field1 bool
		Field2 string
	}

	testCases := []struct {
		name           string
		success        bool
		args           []reflect.Type
		params         interface{}
		expectedResult []interface{}
		expectedError  error
	}{
		{
			name:           "no arguments",
			success:        true,
			args:           []reflect.Type{},
			params:         nil,
			expectedResult: []interface{}{},
			expectedError:  nil,
		},
		{
			name:           "parse one arg : string",
			success:        true,
			args:           []reflect.Type{reflect.TypeOf("")},
			params:         "foo",
			expectedResult: []interface{}{"foo"},
			expectedError:  nil,
		},
		{
			name:           "parse one arg : int",
			success:        true,
			args:           []reflect.Type{reflect.TypeOf(4)},
			params:         4,
			expectedResult: []interface{}{4},
			expectedError:  nil,
		},
		{
			name:           "parse one arg : boolean",
			success:        true,
			args:           []reflect.Type{reflect.TypeOf(true)},
			params:         false,
			expectedResult: []interface{}{false},
			expectedError:  nil,
		},
		{
			name:           "parse one arg : float",
			success:        true,
			args:           []reflect.Type{reflect.TypeOf(float64(2))},
			params:         float64(2),
			expectedResult: []interface{}{float64(2)},
			expectedError:  nil,
		},
		{
			name:           "parse one arg : object",
			success:        true,
			args:           []reflect.Type{reflect.TypeOf(struct{ Foo string }{Foo: ""})},
			params:         struct{ Foo string }{Foo: "foo"},
			expectedResult: []interface{}{struct{ Foo string }{Foo: "foo"}},
			expectedError:  nil,
		},
		{
			name:           "parse one arg : array",
			success:        true,
			args:           []reflect.Type{reflect.TypeOf([]int{0})},
			params:         []int{1, 2, 3},
			expectedResult: []interface{}{[]int{1, 2, 3}},
			expectedError:  nil,
		},
		{
			name:           "parse on arg : type do not match",
			success:        false,
			args:           []reflect.Type{reflect.TypeOf(false)},
			params:         []interface{}{"test"},
			expectedResult: nil,
			expectedError:  ErrInvalidArgType,
		},
		{
			name:           "parse multi arg : int",
			success:        true,
			args:           []reflect.Type{reflect.TypeOf(0), reflect.TypeOf(0)},
			params:         []int{1, 2},
			expectedResult: []interface{}{1, 2},
			expectedError:  nil,
		},
		{
			name:           "parse multi arg : string",
			success:        true,
			args:           []reflect.Type{reflect.TypeOf(""), reflect.TypeOf("")},
			params:         []string{"foo", "bar"},
			expectedResult: []interface{}{"foo", "bar"},
			expectedError:  nil,
		},
		{
			name:           "parse multi arg : boolean",
			success:        true,
			args:           []reflect.Type{reflect.TypeOf(false), reflect.TypeOf(false)},
			params:         []bool{false, true},
			expectedResult: []interface{}{false, true},
			expectedError:  nil,
		},
		{
			name:           "parse multi arg : mix primitive type",
			success:        true,
			args:           []reflect.Type{reflect.TypeOf(false), reflect.TypeOf(""), reflect.TypeOf(0)},
			params:         []interface{}{true, "foo", 5},
			expectedResult: []interface{}{true, "foo", 5},
			expectedError:  nil,
		},
		{
			name:           "parse multi arg : mix primitive type with array",
			success:        true,
			args:           []reflect.Type{reflect.TypeOf(false), reflect.TypeOf(""), reflect.TypeOf([]int{0})},
			params:         []interface{}{true, "foo", []int{1, 2, 3}},
			expectedResult: []interface{}{true, "foo", []int{1, 2, 3}},
			expectedError:  nil,
		},
		{
			name:           "parse multi arg : mix primitive type with array and object",
			success:        true,
			args:           []reflect.Type{reflect.TypeOf(false), reflect.TypeOf(""), reflect.TypeOf([]int{0}), reflect.TypeOf(struct{ Foo string }{Foo: ""})},
			params:         []interface{}{true, "foo", []int{1, 2, 3}, struct{ Foo string }{Foo: "foo"}},
			expectedResult: []interface{}{true, "foo", []int{1, 2, 3}, struct{ Foo string }{Foo: "foo"}},
			expectedError:  nil,
		},
		{
			name:           "parse multi arg : mix primitive type with array and object",
			success:        true,
			args:           []reflect.Type{reflect.TypeOf(false), reflect.TypeOf(""), reflect.TypeOf([]int{0}), reflect.TypeOf(FakeStruct{})},
			params:         []interface{}{true, "foo", []int{1, 2, 3}, FakeStruct{0, true, "struct"}},
			expectedResult: []interface{}{true, "foo", []int{1, 2, 3}, FakeStruct{0, true, "struct"}},
			expectedError:  nil,
		},
		{
			name:           "parse multi arg : type do not match",
			success:        false,
			args:           []reflect.Type{reflect.TypeOf(false), reflect.TypeOf("")},
			params:         []interface{}{true, 4},
			expectedResult: nil,
			expectedError:  ErrInvalidArgType,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			res, err := Arguments(tt.args, tt.params)

			assert.Equal(t, tt.expectedResult, res)
			assert.Equal(t, tt.expectedError, err)
		})
	}
}
