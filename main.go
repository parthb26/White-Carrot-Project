package main

import (
	"net/http"
	"Whit-Carrot-Assignment/handlers"
)

func main() {
	http.HandleFunc("/", handlers.HomeHandler)
	http.HandleFunc("/login", handlers.LoginHandler)
	http.HandleFunc("/callback", handlers.CallbackHandler)

	http.ListenAndServe(":8080", nil)
}
