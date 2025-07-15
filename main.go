package main

import (
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/ws", handleWebSocket)
	log.Fatal(http.ListenAndServe(":4000", nil))
}
