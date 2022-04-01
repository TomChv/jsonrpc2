package server

import (
	"encoding/json"
	"net/http"
)

func (s *JsonRPC2) sendBatch(res []*Response, w http.ResponseWriter) error {
	data, err := json.Marshal(res)
	if err != nil {
		return err
	}

	_, err = w.Write(data)
	if err != nil {
		return err
	}
	return nil
}
