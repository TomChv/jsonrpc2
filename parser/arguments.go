package parser

import (
	"encoding/json"
	"errors"
	"reflect"
)

var (
	ErrInvalidArgsNumber       = errors.New("invalid number of arguments")
	ErrInvalidArgExpectedSlice = errors.New("invalid argument, expected slice")
	ErrInvalidArgType          = errors.New("invalid arg type")
)

// Arguments convert any param into types send in args
//   If no args         -> return empty
//   If 1 arg           -> directly parse the param and return it a single
//   IF 2 or more arg   -> verify that param is an array and loop through it to
//  convert it to an array of interface with correct type
func Arguments(args []reflect.Type, param interface{}) ([]interface{}, error) {
	switch len(args) {
	case 0:
		return []interface{}{}, nil
	case 1:
		p, err := parseArgument(args[0], param)
		if err != nil {
			return nil, err
		}

		return []interface{}{p}, err
	default:
		if reflect.TypeOf(param).Kind() != reflect.Slice {
			return nil, ErrInvalidArgsNumber
		}

		params, err := convertInterfaceToArray(param)
		if err != nil {
			return nil, err
		}

		res := make([]interface{}, len(args))
		for i, e := range params {
			p, err := parseArgument(args[i], e)
			if err != nil {
				return nil, err
			}
			res[i] = p
		}

		return res, nil
	}
}

// parseArgument convert the param into the type of the arg
// Since a simple reflect is not enough to verify if the param is type of arg
// this function use json.Unmarshal to correctly convert the param
func parseArgument(arg reflect.Type, param interface{}) (interface{}, error) {
	expectedType := reflect.StructOf([]reflect.StructField{{
		Name: "Placeholder",
		Type: arg,
	}})
	expected := reflect.New(expectedType).Interface()

	placeholder := struct {
		Placeholder interface{}
	}{
		Placeholder: param,
	}

	data, err := json.Marshal(placeholder)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &expected)
	if err != nil {
		return nil, ErrInvalidArgType
	}

	value := reflect.ValueOf(expected).Elem().FieldByName("Placeholder")
	return value.Interface(), nil
}

// convertInterfaceToArray is a utility function used to transform
// an interface into an array of interface
// The result can then be used to populate arguments to the dispatcher
func convertInterfaceToArray(value interface{}) ([]interface{}, error) {
	var out []interface{}

	reflectValue := reflect.ValueOf(value)
	if reflectValue.Kind() != reflect.Slice {
		return nil, ErrInvalidArgExpectedSlice
	}

	for i := 0; i < reflectValue.Len(); i++ {
		out = append(out, reflectValue.Index(i).Interface())
	}

	return out, nil
}
