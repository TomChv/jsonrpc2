package parser

import (
	"encoding/json"

	"github.com/TomChv/jsonrpc2/common"
	"github.com/TomChv/jsonrpc2/validator"
)

// Batch parse an array of byte to convert it into an array of request and
// errors
// This function do not fail if any request isn't valid, it will add it to
// errors array
// To differentiate parsing error from invalid request, boolean is sent as
// first return value
func Batch(body []byte) (bool, []*common.Request, []error) {
	var reqs []common.Request

	if err := json.Unmarshal(body, &reqs); err != nil {
		return false, nil, nil
	}

	res := []*common.Request{}
	var errs []error

	for _, req := range reqs {
		req := req
		if err := validator.JsonRPCRequest(&req); err != nil {
			errs = append(errs, err)
			continue
		}
		// Have an issue if append with &req
		res = append(res, &req)
	}

	return true, res, errs
}
