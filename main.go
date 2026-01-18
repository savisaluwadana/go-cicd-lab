package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"
)

// Book represents a library book
type Book struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author"`
	Year   int    `json:"year"`
}

// in-memory store (safe for concurrent access)
var (
	store = struct {
		sync.RWMutex
		m map[string]Book
	}{m: make(map[string]Book)}
)

func listBooks(w http.ResponseWriter, r *http.Request) {
	store.RLock()
	defer store.RUnlock()
	books := make([]Book, 0, len(store.m))
	for _, b := range store.m {
		books = append(books, b)
	}
	jsonResponse(w, books, http.StatusOK)
}

func getBook(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}
	store.RLock()
	b, ok := store.m[id]
	store.RUnlock()
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	jsonResponse(w, b, http.StatusOK)
}

func createBook(w http.ResponseWriter, r *http.Request) {
	var b Book
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if b.ID == "" {
		http.Error(w, "id required", http.StatusBadRequest)
		return
	}
	store.Lock()
	store.m[b.ID] = b
	store.Unlock()
	jsonResponse(w, b, http.StatusCreated)
}

func updateBook(w http.ResponseWriter, r *http.Request) {
	var b Book
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if b.ID == "" {
		http.Error(w, "id required", http.StatusBadRequest)
		return
	}
	store.Lock()
	_, exists := store.m[b.ID]
	if !exists {
		store.Unlock()
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	store.m[b.ID] = b
	store.Unlock()
	jsonResponse(w, b, http.StatusOK)
}

func deleteBook(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}
	store.Lock()
	defer store.Unlock()
	if _, ok := store.m[id]; !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	delete(store.m, id)
	w.WriteHeader(http.StatusNoContent)
}

func jsonResponse(w http.ResponseWriter, v interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func main() {
	// seed a couple books
	store.Lock()
	store.m["1"] = Book{ID: "1", Title: "The Go Programming Language", Author: "Alan A. A. Donovan", Year: 2015}
	store.m["2"] = Book{ID: "2", Title: "Clean Code", Author: "Robert C. Martin", Year: 2008}
	store.Unlock()

	mux := http.NewServeMux()
	// API
	mux.HandleFunc("/api/books", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			listBooks(w, r)
		case http.MethodPost:
			createBook(w, r)
		case http.MethodPut:
			updateBook(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/book", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getBook(w, r)
		case http.MethodDelete:
			deleteBook(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Serve frontend static files from ./frontend
	fs := http.FileServer(http.Dir("./frontend"))
	mux.Handle("/", fs)

	addr := ":8080"
	if p := os.Getenv("PORT"); p != "" {
		addr = ":" + p
	}
	log.Printf("listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
