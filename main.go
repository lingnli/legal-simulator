package main

import (
	"fmt"
	"net/http"
	"os"
)

func webhook(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "WEBHOOK OK")
}

func home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "HOME OK")
}

func main() {

	http.HandleFunc("/", home)
	http.HandleFunc("/webhook", webhook)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Println("start :", port)

	http.ListenAndServe(":"+port, nil)
}
