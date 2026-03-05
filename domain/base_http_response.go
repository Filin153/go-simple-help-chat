package domain

type BaseResponse[T any] struct {
	Msg   string `json:"msg"`
	Error bool   `json:"error"`
	Data  *T     `json:"data,omitempty"`
}
