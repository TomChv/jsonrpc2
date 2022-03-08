package server

import (
	"errors"
	"reflect"
)

var (
	ErrInvalidServiceProcedures = errors.New("services procedures does not match ServiceProcedure type")
)

// validateService ensure that each public methods of the service is compliant
// with the type : (interface{}, error).
//
// It uses reflect to loop though each public methods of the service
// Then it verifies that method has the same return type than expected.
//
// If service is compliant, validateService return true, else false.
func validateService(service interface{}) bool {
	st := reflect.TypeOf(service)

	for i := 0; i < st.NumMethod(); i++ {
		if !st.Method(i).IsExported() {
			continue
		}

		if st.Method(i).Func.Type().NumOut() != 2 {
			return false
		}

		if !st.Method(i).Func.Type().Out(1).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			return false
		}
	}
	return true
}
