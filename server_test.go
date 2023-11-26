package webh_test

import (
	"github.com/tomiok/webh"
	"net/http"
	"syscall"
	"testing"
	"time"
)

func Test_serverCreate(t *testing.T) {
	s := webh.NewServer("8080", webh.WithLogger("hello"), webh.WithHeartbeat("/ping"))

	s.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("hello"))
	})

	go func() {
		client := http.Client{}
		_, err := client.Get("http://localhost:8080/test")
		if err != nil {
			panic(err)
		}

		//fmt.Println(fmt.Sprintf("%+v", res.Body))
		time.Sleep(5 * time.Second)
		_ = syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	}()

	s.Start()
}
