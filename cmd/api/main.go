package main

import (
	"log"
	"net/http"

	"github.com/cezartdev/go_rate_limiter/internal/config"
)

func main() {

	config.LoadEnv()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello world"))
	})

	addr := ":8080"

	log.Printf("Listening on http://localhost%s", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
