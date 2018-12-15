package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func main() {
	fmt.Printf("starting_server: :8080\n")
	http.HandleFunc("/pong", func(w http.ResponseWriter, r *http.Request) {
		_, err := ioutil.ReadAll(r.Body)
		if err == nil {
			r.Body.Close()
		}
		fmt.Fprintf(w, "PONG")
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}
