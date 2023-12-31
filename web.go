package webh

import (
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type HttpResponse struct {
	Message   string `json:"message"`
	Data      any    `json:"data,omitempty"`
	Success   bool   `json:"success"`
	RequestID string `json:"request_id,omitempty"`
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

// DJson is a generic way to unmarshal your JSON.
func DJson[t any](b io.ReadCloser, target *t) (*t, error) {
	defer func() {
		_ = b.Close()
	}()

	err := json.NewDecoder(b).Decode(target)

	if err != nil {
		return nil, err
	}
	return target, nil
}

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

// Res is a wrapper for web handlers that are compliant with the stdlib signature but returning an error.
func Res(status int, w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(data)
}

// ResErr is a wrapper for web handlers that are compliant with the stdlib signature but returning an error.
func ResErr(status int, w http.ResponseWriter, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(data)
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

// FileServer will be used to serve the static files, please specify the route.
func (s *Server) FileServer(path string) {
	workDir, _ := os.Getwd()
	filesDir := http.Dir(filepath.Join(workDir, path))
	fs(s.Mux, "/"+path, filesDir)
}

// fs conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
func fs(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("file server does not permit any URL parameters")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", http.StatusMovedPermanently).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
}
