package dto

type WebResponse[T any] struct {
	Code    int    `json:"code"             example:"200"`
	Message string `json:"message"          example:"Success"`
	Data    T      `json:"data,omitempty"`
	Errors  any    `json:"errors,omitempty"`
}
