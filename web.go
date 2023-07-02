package webh

import (
	"encoding/json"
	"errors"
	"net/http"
)

type HttpResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Success bool        `json:"success"`
}

type ErrHTTP struct {
	Message string
	Code    int
}

func (e ErrHTTP) Error() string {
	return e.Message
}

func wrapErrorResponse(w http.ResponseWriter, err error) {
	_err := transform(err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(_err.Code)

	r, _ := json.Marshal(HttpResponse{
		Message: _err.Message,
		Success: false,
	})
	_, _ = w.Write(r)
}

func transform(e error) *ErrHTTP {
	var webErr ErrHTTP
	if errors.As(e, &webErr) {
		err := e.(ErrHTTP)
		return &err
	}

	return &ErrHTTP{
		Message: "errors during the request",
		Code:    http.StatusInternalServerError,
	}
}
