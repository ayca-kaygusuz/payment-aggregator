package database

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

// StartServer starts the HTTP server and handles the /callback route
func StartServer(db *Database) {
	// Handle /callback route
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		HandleCallback(w, r, db)
	})

	// Log the server start and errors
	// TODO: adjust env variables and take these dynamically
	log.Println("Server is starting on :8080")
	log.Fatal(http.ListenAndServe("localhost:8080", nil))
}

func HandleCallback(w http.ResponseWriter, r *http.Request, db *Database) {

	// Only allow POST requests
	if r.Method != http.MethodPost {
		log.Println("Ignored non-POST request:", r.Method)
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read the body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("Error reading body:", err)
		http.Error(w, "cannot read body", http.StatusBadRequest)
		return
	}

	// Log raw body
	log.Println("Raw Body:", string(body))

	// Try to parse and pretty-print JSON (optional, for structured APIs)
	var parsed map[string]interface{}
	if err := json.Unmarshal(body, &parsed); err == nil {
		pretty, _ := json.MarshalIndent(parsed, "", "  ")
		log.Println("Parsed JSON:")
		log.Println(string(pretty))
	} else {
		log.Println("Body is not valid JSON")
	}

	// Respond OK
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Callback received"))
}
