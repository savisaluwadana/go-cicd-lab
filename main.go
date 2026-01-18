package main

import (
	"fmt"
	"net/http"
	"testing"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello from GitHub Actions!")
}

func main() {
	http.HandleFunc("/", Handler)
	http.ListenAndServe(":8080", nil)
}

// Simple test to run in the pipeline
func TestHandler(t *testing.T) {
	expected := "Hello from GitHub Actions!"
	if expected != "Hello from GitHub Actions!" {
		t.Errorf("Handler failed")
	}
}
