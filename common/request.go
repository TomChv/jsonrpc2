package common

// RequestParam is a Structured value that holds the parameter values to be used
// during the invocation of the method
type RequestParam = interface{}

// Request is a JSON-RPC 2.0 Object that represent a call to the server
type Request struct {
	// JsonRPC is the version of JSON-RPC protocol, must be "2.0"
	JsonRpc *string `json:"jsonrpc,required"`

	// Method containing the name of the method to be invoked.
	Method *string `json:"method,required"`

	// Params is a Structured value that holds the parameter values to be used
	// during the invocation of the method.
	Params RequestParam `json:"params,omitempty"`

	// ID is an identifier established by the Client that must contain a String,
	// Number, or NULL value if included.
	// If it is not included it is assumed to be a notification.
	ID RequestID `json:"id,omitempty"`
}
