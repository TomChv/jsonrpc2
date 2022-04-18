package parser

import (
	"encoding/json"
	"errors"
)

var ErrEmptyBatch = errors.New("empty batch")

// Batch parse an array of byte to convert it as an array of raw request
func Batch(body []byte) ([][]byte, error) {
	var reqs []interface{}
	if err := json.Unmarshal(body, &reqs); err != nil {
		return nil, err
	}

	if len(reqs) == 0 {
		return nil, ErrEmptyBatch
	}

	var res [][]byte
	for _, req := range reqs {
		d, err := json.Marshal(req)
		if err != nil {
			return nil, err
		}
		res = append(res, d)
	}

	return res, nil
}
