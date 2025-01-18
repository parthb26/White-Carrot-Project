package main

import (
	"log"
	"net/http"

	"github.com/parthb26/WhiteCarrot-Project/handlers"
)

func main() {
	// Initialize the Google OAuth configuration with your credentials
	handlers.InitGoogleOAuth("your-client-id", "your-client-secret")

	// Set up the routes and their handlers
	http.HandleFunc("/login", handlers.LoginHandler)
	http.HandleFunc("/callback", handlers.CallbackHandler)

	// Starting server 
	log.Println("Server starting on http://localhost:8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe error: ", err)
	}
}
