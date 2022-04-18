package common

import (
	"encoding/json"
	"net/http"
	"sync"
)

type Batch struct {
	responses []*Response
	l         sync.Mutex
}

func (b *Batch) Get() []*Response {
	return b.responses
}

func (b *Batch) Append(res *Response) {
	b.l.Lock()
	defer b.l.Unlock()

	b.responses = append(b.responses, res)
}

func (b *Batch) Send(w http.ResponseWriter) error {
	data, err := json.Marshal(b.responses)
	if err != nil {
		return err
	}

	_, err = w.Write(data)
	if err != nil {
		return err
	}
	return nil
}
