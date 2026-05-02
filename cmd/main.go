// To run curl -i -X POST http://localhost:8080/ -H "Content-Type: text/plain" -d "https://lessgo.ru"
package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

var storage = make(map[string]string)

func main() {
	run()
}

func run() {
	http.HandleFunc("/", shortenHandler)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func shortenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if r.Header.Get("Content-Type") != "text/plain" {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	originalURL := strings.TrimSpace(string(body))
	if originalURL == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	shortKey := generateShortKey()
	storage[shortKey] = originalURL

	shortURL := fmt.Sprintf("http://localhost:8080/%s", shortKey)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))
}

func generateShortKey() string {
	bytes := make([]byte, 4)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)
}
