[![License](https://img.shields.io/github/license/tomiok/webh?style=for-the-badge)](https://github.com/tomiok/webh/blob/master/LICENSE)

# webh, the web helper for Golang

## Include web code snippets for speed up the development.


This is a tiny yet powerful library was built for do not write every single
time the web server.

We provide several capabilities used in the industry like
middlewares, graceful shutdown, logging, ability to unwrap web handlers that return an
error in order to avoid weird returns.

A clean mechanism to log errors among the http requests.

Encoder and Decoder for JSON.

A custom error type for web purposes.

The Server is created on top of [chi](https://go-chi.io) web library, so will have all
the same features and ways to declare endpoints.


Heartbeat at `/ping` is already declared for you, among recover and logger. Do not worry to add them.

### Installation

```shell
go get -u github.com/tomiok/webh
```

### Create and start the server with one endpoint.
```go
s := webh.NewServer("8080", webh.WithLogger("hello"), webh.WithHeartbeat("/ping"))

s.Get("/hello", func(w http.ResponseWriter, r *http.Request){
	//.....
})

s.Start()
```

or use a custom handler returning an error.

```go
package web

import (
	"fmt"
	"net/http"
)

func HelloHandler(w http.ResponseWriter, r *http.Request) error {
	_, err := fmt.Fprint(w, "hello")
	return err
}
```
```go
package main

import "github.com/tomiok/webh"

func main() {
	s := webh.NewServer("8080", webh.WithLogger("hello"), webh.WithHeartbeat("/ping"))

	s.Get("/hello", webh.Unwrap(HelloHandler))

	s.Start()
}
```