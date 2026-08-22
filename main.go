package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func webhook(w http.ResponseWriter, r *http.Request) {
	fmt.Println("=== WEBHOOK HIT ===")

	body, _ := io.ReadAll(r.Body)

	fmt.Println(string(body))

	fmt.Fprint(w, "OK")
}

func home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "HOME OK")
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
