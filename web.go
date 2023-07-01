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

type errHTTP struct {
	Message string
	Code    int
}

func (e errHTTP) Error() string {
	return e.Message
}

func transform(e error) *errHTTP {
	var webErr errHTTP
	if errors.As(e, &webErr) {
		err := e.(errHTTP)
		return &err
	}

	return &errHTTP{
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
