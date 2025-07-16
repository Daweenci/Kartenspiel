package main

import (
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "client/index.html")
	})
	http.HandleFunc("/ws", handleWebSocket)
	log.Fatal(http.ListenAndServe(":4000", nil))
}
