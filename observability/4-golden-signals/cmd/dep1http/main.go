package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	fmt.Printf("starting_server: :8080\n")
	http.HandleFunc("/pong", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "PONG")
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}
