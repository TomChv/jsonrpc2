package server

import "github.com/TomChv/jsonrpc2/common"

// ParsingError when invalid JSON was received by the server
func ParsingError(err error) *common.RpcError {
	return &common.RpcError{
		Code:    -32700,
		Message: "Parse error",
		Data:    err.Error(),
	}
}

// InvalidRequestError when the JSON sent is not a valid Request object
func InvalidRequestError(err error) *common.RpcError {
	return &common.RpcError{
		Code:    -32600,
		Message: "Invalid Request",
		Data:    err.Error(),
	}
}

// MethodNotFoundError when the method does not exist / is not available
func MethodNotFoundError(err error) *common.RpcError {
	return &common.RpcError{
		Code:    -32601,
		Message: "Method not found",
		Data:    err.Error(),
	}
}

// InvalidParamsError for invalid method parameter(s)
func InvalidParamsError(err error) *common.RpcError {
	return &common.RpcError{
		Code:    -32602,
		Message: "Invalid params",
		Data:    err.Error(),
	}
}

// InternalError  for internal JSON-RPC error
func InternalError(err error) *common.RpcError {
	return &common.RpcError{
		Code:    -32603,
		Message: "Internal error",
		Data:    err.Error(),
	}
}

// CustomError reserved for implementation-defined server-errors
// Code must be bound with the range -32000 and -32099 according to official
// JSON-RPC 2.0 specification.
func CustomError(code int64, err error) *common.RpcError {
	return &common.RpcError{
		Code:    code,
		Message: "Server error",
		Data:    err.Error(),
	}
}
