package webh

import (
	"net/http"
	"syscall"
	"testing"
	"time"
)

func Test_serverCreate(t *testing.T) {
	s := NewServer("8080", WithLogger("hello"), WithHeartbeat("/ping"))

	s.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("hello"))
	})

	go func() {
		time.Sleep(5 * time.Second)
		_ = syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	}()
	s.Start()
}
