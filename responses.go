package webh

import (
	"encoding/json"
	"net/http"
)

// Response is a wrapper for web handlers that are compliant with the stdlib.
func Response(status int, w http.ResponseWriter, msg string, data any) {
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(HttpResponse{
		Message: msg,
		Data:    data,
		Success: status < 399,
	})
}

// ResponseErr is a wrapper for web handlers that are compliant with the stdlib signature but returning an error.
func ResponseErr(status int, w http.ResponseWriter, msg string, data any) error {
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(HttpResponse{
		Message: msg,
		Data:    data,
		Success: status < 399,
	})
}
