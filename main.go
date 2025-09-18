package main

import (
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "client/index.html")
	})
	http.HandleFunc("/login", handleLogin)
	http.HandleFunc("/register", handleRegister)
	http.HandleFunc("/ws", handleWebSocket)
	log.Println("WebSocket server started on :4000")
	log.Fatal(http.ListenAndServe(":4000", nil))
}
