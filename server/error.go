package server

import "github.com/TomChv/jsonrpc2/common"

// ParsingError when invalid JSON was received by the server
func ParsingError(data interface{}) *common.RpcError {
	return &common.RpcError{
		Code:    -32700,
		Message: "Parse error",
		Data:    data,
	}
}

// InvalidRequestError when the JSON sent is not a valid Request object
func InvalidRequestError(data interface{}) *common.RpcError {
	return &common.RpcError{
		Code:    -32600,
		Message: "Invalid Request",
		Data:    data,
	}
}

// MethodNotFoundError when the method does not exist / is not available
func MethodNotFoundError(data interface{}) *common.RpcError {
	return &common.RpcError{
		Code:    -32601,
		Message: "Method not found",
		Data:    data,
	}
}

// InvalidParamsError for invalid method parameter(s)
func InvalidParamsError(data interface{}) *common.RpcError {
	return &common.RpcError{
		Code:    -32602,
		Message: "Invalid params",
		Data:    data,
	}
}

// InternalError  for internal JSON-RPC error
func InternalError(data interface{}) *common.RpcError {
	return &common.RpcError{
		Code:    -32603,
		Message: "Internal error",
		Data:    data,
	}
}

// CustomError reserved for implementation-defined server-errors
// Code must be bound with the range -32000 and -32099 according to official
// JSON-RPC 2.0 specification.
func CustomError(code int64, data interface{}) *common.RpcError {
	return &common.RpcError{
		Code:    code,
		Message: "Server error",
		Data:    data,
	}
}
