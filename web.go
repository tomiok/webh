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

type Err struct {
	Message string
	Code    int
}

func (e Err) Error() string {
	return e.Message
}

func transform(e error) *Err {
	var webErr *Err
	if errors.As(e, webErr) {
		err := e.(Err)
		return &err
	}

	return &Err{
		Message: "errors during the request",
		Code:    http.StatusInternalServerError,
	}
}

func ReturnErr(w http.ResponseWriter, err error) {
	_err := transform(err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(_err.Code)

	r, _ := json.Marshal(HttpResponse{
		Message: _err.Message,
		Success: false,
	})
	_, _ = w.Write(r)
}

func ResponseUnauthorized(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	r, _ := json.Marshal(HttpResponse{
		Message: msg,
		Success: false,
	})
	_, _ = w.Write(r)
}

func ResponseOK(w http.ResponseWriter, msg string, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(HttpResponse{
		Message: msg,
		Success: true,
		Data:    data,
	})
	return nil
}

func ResponseCreated(w http.ResponseWriter, msg string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(HttpResponse{
		Message: msg,
		Success: true,
		Data:    data,
	})
}

func ResponseNoContent(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}
