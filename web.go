package webh

import (
	"encoding/json"
	"errors"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
)

type HttpResponse struct {
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Success   bool        `json:"success"`
	RequestID string      `json:"request_id,omitempty"`
}

type ErrHTTP struct {
	Message string
	Code    int
}

func (e ErrHTTP) Error() string {
	return e.Message
}

func wrapErrorResponse(w http.ResponseWriter, requestID string, err error) {
	_err := transform(err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(_err.Code)

	r, _ := json.Marshal(HttpResponse{
		Message: _err.Message,
		Success: false,

		RequestID: requestID,
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

// EJson is the shorthand to encode JSON.
func EJson(w io.Writer, v any) error {
	return json.NewEncoder(w).Encode(v)
}

// Djson is a generic way to unmarshal your JSON.
func Djson[t any](b io.ReadCloser, target *t) (*t, error) {
	defer func() {
		log.Info().Msg("closing")
		_ = b.Close()
	}()

	err := json.NewDecoder(b).Decode(target)

	if err != nil {
		return nil, err
	}
	return target, nil
}
