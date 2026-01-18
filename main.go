package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"

	_ "modernc.org/sqlite"
)

// Book represents a library book
type Book struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author"`
	Year   int    `json:"year"`
}

var db *sql.DB

func initDB(path string) error {
	var err error
	db, err = sql.Open("sqlite", path)
	if err != nil {
		return err
	}
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS books (
		id TEXT PRIMARY KEY,
		title TEXT,
		author TEXT,
		year INTEGER
	);`)
	return err
}

func listBooks(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id,title,author,year FROM books")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	var books []Book
	for rows.Next() {
		var b Book
		if err := rows.Scan(&b.ID, &b.Title, &b.Author, &b.Year); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
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
	var b Book
	err := db.QueryRow("SELECT id,title,author,year FROM books WHERE id = ?", id).Scan(&b.ID, &b.Title, &b.Author, &b.Year)
	if err == sql.ErrNoRows {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
	_, err := db.Exec("INSERT INTO books(id,title,author,year) VALUES(?,?,?,?)", b.ID, b.Title, b.Author, b.Year)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
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
	res, err := db.Exec("UPDATE books SET title=?,author=?,year=? WHERE id=?", b.Title, b.Author, b.Year, b.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	jsonResponse(w, b, http.StatusOK)
}

func deleteBook(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}
	res, err := db.Exec("DELETE FROM books WHERE id = ?", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func jsonResponse(w http.ResponseWriter, v interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "data.db"
	}
	if err := initDB(dbPath); err != nil {
		log.Fatalf("db init: %v", err)
	}

	// seed data if empty
	var count int
	_ = db.QueryRow("SELECT COUNT(1) FROM books").Scan(&count)
	if count == 0 {
		_, _ = db.Exec("INSERT INTO books(id,title,author,year) VALUES(?,?,?,?)", "1", "The Go Programming Language", "Alan A. A. Donovan", 2015)
		_, _ = db.Exec("INSERT INTO books(id,title,author,year) VALUES(?,?,?,?)", "2", "Clean Code", "Robert C. Martin", 2008)
	}

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
	log.Printf("listening on %s (DB=%s)", addr, dbPath)
	log.Fatal(http.ListenAndServe(addr, mux))
}
