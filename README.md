# go-cicd-lab — Library Manager

Simple Go backend providing a small CRUD API for managing books and a minimal HTML/CSS/JS frontend.

Run locally:

```bash
go run .
# open http://localhost:8080
```

Build & run with Docker:

```bash
docker build -t youruser/go-cicd-lab:local .
docker run -p 8080:8080 youruser/go-cicd-lab:local
```

Run with Docker Compose (SQLite data persisted in named volume):

```bash
# Build and start services
docker compose up --build -d

# Open the app at http://localhost:8080

# Inspect the DB file using the sqlite helper container:
docker compose run --rm sqlite sqlite3 /data/data.db "SELECT * FROM books;"

# To stop remove containers and volume (data will be kept unless you remove the volume):
docker compose down
```

API endpoints:
- `GET /api/books` — list all books
- `GET /api/book?id={id}` — get book by id
- `POST /api/books` — create book (JSON body)
- `PUT /api/books` — update book (JSON body)
- `DELETE /api/book?id={id}` — delete book
# go-cicd-lab