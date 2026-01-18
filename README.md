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

API endpoints:
- `GET /api/books` — list all books
- `GET /api/book?id={id}` — get book by id
- `POST /api/books` — create book (JSON body)
- `PUT /api/books` — update book (JSON body)
- `DELETE /api/book?id={id}` — delete book
# go-cicd-lab