package parser

import (
	"encoding/json"

	"github.com/TomChv/jsonrpc2/common"
)

// Batch parse an array of byte to convert it into an array of request
func Batch(body []byte) ([]*common.Request, error) {
	var reqs []*common.Request

	if err := json.Unmarshal(body, &reqs); err != nil {
		return nil, err
	}

	return reqs, nil
}
