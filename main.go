package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/", handlerFunc)
	fmt.Println("Starting the server on :3000...")
	_ = http.ListenAndServe(":3000", nil)
}

func handlerFunc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, "<h1>Welcome to my awesome site</h1>")
}

func handlerFunc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, "<h1>Welcome to my awesome site</h1>")
}
