package main

import (
	"log"
	"net/http"

	"example.com/profile-service/internal/handler/http"
)

func main() {
	http.HandleFunc("/profile", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			http.CreateProfileHandler(w, r)
		case http.MethodGet:
			http.GetProfileHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	log.Println("Server starting on port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("could not start server: %v\n", err)
	}
}
