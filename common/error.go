package common

import "fmt"

// RpcError is a JSON-RPC 2.0 Object that represent an error from the server
type RpcError struct {
	// Code is a number that indicate the error type that occurred
	Code int64

	// Message is a short description of the error
	Message string

	// Data contains additional information about the error
	Data interface{}
}

func (e *RpcError) Error() string {
	return fmt.Sprintf("{\n"+
		"\tCode: %v\n"+
		"\tMessage: %v\n"+
		"\tData: %v\n"+
		"}", e.Code, e.Message, e.Data)
}
