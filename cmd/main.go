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
	"sync"
)

type urlStorage struct {
	mu   sync.RWMutex
	urls map[string]string
}

func newURLStorage() *urlStorage {
	return &urlStorage{
		urls: make(map[string]string),
	}
}

func (s *urlStorage) save(id, originalURL string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.urls[id] = originalURL
}

func (s *urlStorage) get(id string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	originalURL, ok := s.urls[id]
	return originalURL, ok
}

var storage = newURLStorage()

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

const (
	serverAddr = ":8080"
	baseURL    = "http://localhost:8080"
)

func run() error {
	http.HandleFunc("/", handleRoot)
	return http.ListenAndServe(serverAddr, nil)
}

// handleRoot routes requests based on the method:
// POST / - shorten a URL, GET /{id} - return a redirect to the original
func handleRoot(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		handleShorten(w, r)
	case http.MethodGet:
		handleRedirect(w, r)
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}

func handleShorten(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	originalURL := strings.TrimSpace(string(body))
	if originalURL == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id := generateShortID()
	storage.save(id, originalURL)

	shortURL := fmt.Sprintf("%s/%s", baseURL, id)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))
}

func handleRedirect(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	originalURL, ok := storage.get(id)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func generateShortID() string {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}
