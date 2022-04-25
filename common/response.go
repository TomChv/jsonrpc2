package common

// ResponseResult is required on success, this member must not exist if there
// was an error invoking the method
type ResponseResult = interface{}

// Response is a JSON-RPC 2.0 Object that represent a response from the server
type Response struct {
	// JsonRPC is the version of JSON-RPC protocol, must be "2.0"
	JsonRpc string `json:"jsonrpc" binding:"register"`

	// Result is required on success, this member must not exist if there was
	// an error invoking the method
	Result ResponseResult `json:"result,omitempty"`

	// Error is required on error, this member must not exist if there was
	// no error triggered during invocation
	Error *RpcError `json:"error,omitempty"`

	// ID is required, it must be the same value as the id in the Request
	// Object.
	ID RequestID `json:"id"`
}
