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

type err struct {
	Message string
	Code    int
}

func (e err) Error() string {
	return e.Message
}

func transform(e error) *err {
	var webErr *err
	if errors.As(e, webErr) {
		err := e.(err)
		return &err
	}

	return &err{
		Message: "errors during the request",
		Code:    http.StatusInternalServerError,
	}
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
