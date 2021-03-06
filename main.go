package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/sinmetal/gaeblob/backend"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Listening on port %s", port)
	http.HandleFunc("/v1/upload", backend.UploadURLHandler)
	http.HandleFunc("/v1/download", backend.DownloadURLHandler)
	http.HandleFunc("/static/", backend.StaticContentsHandler)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), http.DefaultServeMux); err != nil {
		log.Printf("failed ListenAndServe err=%+v", err)
	}
}
