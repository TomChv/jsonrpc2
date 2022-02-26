package server

import (
	"errors"
	"reflect"
)

var (
	ErrInvalidServiceProcedures = errors.New("services procedures does not match ServiceProcedure type")
)

// validateService ensure that each public methods of the service is compliant
// with the JSON RPC 2.0 response type : (Response, RpcError).
//
// It uses reflect to loop though each public methods of the service
// Then it verifies that method has the same return type than expected.
//
// If service is compliant, validateService return true, else false.
func validateService(service interface{}) bool {
	expectedSignatureTypes := []reflect.Type{
		reflect.TypeOf(&Response{}),
		reflect.TypeOf(&RpcError{}),
	}
	st := reflect.TypeOf(service)

	for i := 0; i < st.NumMethod(); i++ {
		if !st.Method(i).IsExported() {
			continue
		}

		if st.Method(i).Func.Type().NumOut() != 2 {
			return false
		}

		for j := 0; j < st.Method(i).Func.Type().NumOut(); j++ {
			if !st.Method(i).Func.Type().Out(j).AssignableTo(expectedSignatureTypes[j]) {
				return false
			}
		}
	}
	return true
}
