package main

import "net/http"

func GetDefaultServer(handler http.Handler) *http.Server {
	return &http.Server{Addr: "127.0.0.1:8000", Handler: handler}
}
