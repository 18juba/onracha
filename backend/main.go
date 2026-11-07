package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if err := json.NewEncoder(w).Encode(map[string]string{
			"message": "Hello, World!",
		}); err != nil {
			log.Printf("write hello response: %v", err)
		}
	})

	log.Println("ONRACHA API listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
