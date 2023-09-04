package main

import (
	"log"
	"net/http"
)

func handleEcho(res http.ResponseWriter, req *http.Request) {
	res.Write([]byte("echo!"))
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/echo", handleEcho)

	err := http.ListenAndServe(":80", mux)
	if err != http.ErrServerClosed {
		log.Fatalf("error running server: %v", err)
	}
}
