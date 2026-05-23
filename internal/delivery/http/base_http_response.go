package http

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type baseResponse[T any] struct {
	Msg   string `json:"msg"`
	Error bool   `json:"error"`
	Data  *T     `json:"data,omitempty"`
}

func SendBaseResponse[T any](w http.ResponseWriter, code int, msg string, isError bool, data *T) error {
	baseResp := baseResponse[T]{
		Msg:   msg,
		Error: isError,
		Data:  data,
	}
	var dataForSend bytes.Buffer
	err := json.NewEncoder(&dataForSend).Encode(baseResp)
	if err != nil {
		return err
	}

	w.WriteHeader(code)
	_, err = w.Write(dataForSend.Bytes())
	if err != nil {
		return err
	}

	return nil
}
