package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
)

// Globales CORS Middleware - wird für ALLE Requests angewendet
func corsHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// CORS Headers für alle Requests setzen
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// OPTIONS Preflight Request behandeln
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Weiter zum eigentlichen Handler
		h.ServeHTTP(w, r)
	})
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("No .env file found, relying on system env vars")
	}

	if err := initDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer closeDB()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Println("Shutting down ...")
		closeDB()
		os.Exit(0)
	}()

	// Einen neuen ServeMux erstellen
	mux := http.NewServeMux()

	// Routes wie gewohnt definieren - OHNE CORS Code!
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "client/index.html")
	})
	mux.HandleFunc("/login", handleLogin)
	mux.HandleFunc("/register", handleRegister)
	mux.HandleFunc("/ws", handleWebSocket)

	port := os.Getenv("PORT")
	if port == "" {
		port = "4000"
	}

	log.Printf("Server started on :%s with CORS enabled", port)

	// WICHTIG: corsHandler wraps den gesamten mux
	log.Fatal(http.ListenAndServe(":"+port, corsHandler(mux)))
}
