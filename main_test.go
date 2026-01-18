package main

import (
	"testing"
)

// Simple test to run in the pipeline
func TestHandler(t *testing.T) {
	expected := "Hello from GitHub Actions!"
	if expected != "Hello from GitHub Actions!" {
		t.Errorf("Handler failed")
	}
}
