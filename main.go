package main

import (
	"fmt"
	"net/http"
	"os"
)

func webhook(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "OK")
}

func main() {

	http.HandleFunc("/webhook", webhook)

	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	fmt.Println("start :", port)

	http.ListenAndServe(":"+port, nil)
}
