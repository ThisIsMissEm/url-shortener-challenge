package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	// For generating URLs
	"crypto/rand"
	"encoding/base64"
)

var storage = struct {
	sync.RWMutex
	urls map[string]string
}{urls: make(map[string]string)}

func main() {
	http.HandleFunc("/new", handleNew)
	http.HandleFunc("/", handleRedirect)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Helper to respond with nicer error messages:
func respondWithError(res http.ResponseWriter, statusCode int) {
	http.Error(res, http.StatusText(statusCode), statusCode)
}

func handleNew(res http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		res.Header().Set("Allow", "GET")
		respondWithError(res, 405)
		return
	}

	url := req.FormValue("url")

	storage.Lock()
	defer storage.Unlock()

	slug, err := generateSlug(8)
	if err != nil {
		respondWithError(res, 500)
		return
	}

	_, ok := storage.urls[slug]
	if ok {
		respondWithError(res, 500)
		return
	}

	storage.urls[slug] = url

	io.WriteString(res, slug)
}

func handleRedirect(res http.ResponseWriter, req *http.Request) {
	if req.URL.Path == "/" {
		fmt.Println(storage.urls)
		respondWithError(res, 501)
		return
	}

	slug := strings.SplitN(req.URL.Path, "/", 2)[1]

	storage.RLock()
	defer storage.RUnlock()

	newUrl, ok := storage.urls[slug]
	if !ok {
		respondWithError(res, 404)
		return
	}

	http.Redirect(res, req, newUrl, 302)
}

func generateRandomBytes(size int) (bytes []byte, err error) {
	buffer := make([]byte, size)
	n, err := rand.Read(buffer)
	_ = n

	if err != nil {
		return nil, err
	}

	return buffer, nil
}

func generateSlug(size int) (slug string, err error) {
	bytes, err := generateRandomBytes(size)
	if err != nil {
		return "", err
	}

	slug = base64.RawURLEncoding.EncodeToString(bytes)
	slug = strings.Replace(slug, "-", "j", -1)
	slug = strings.Replace(slug, "_", "K", -1)
	slug = slug[:size]

	return slug, nil
}
