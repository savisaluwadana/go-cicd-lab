package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

var store = struct {
	sync.Mutex
	m map[string]Book
}{
	m: make(map[string]Book),
}

func TestListAndCreate(t *testing.T) {
	// simple integration test against mux
	mux := http.NewServeMux()
	mux.HandleFunc("/api/books", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			listBooks(w, r)
		case http.MethodPost:
			createBook(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// ensure empty store for test
	store.Lock()
	store.m = make(map[string]Book)
	store.Unlock()

	// create a book
	b := Book{ID: "x1", Title: "T", Author: "A", Year: 2020}
	body, _ := json.Marshal(b)
	req := httptest.NewRequest(http.MethodPost, "/api/books", bytes.NewReader(body))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 got %d", w.Code)
	}

	// list
	req = httptest.NewRequest(http.MethodGet, "/api/books", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", w.Code)
	}
	var books []Book
	if err := json.NewDecoder(w.Body).Decode(&books); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(books) != 1 {
		t.Fatalf("expected 1 book got %d", len(books))
	}
}
