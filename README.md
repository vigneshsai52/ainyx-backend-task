# Ainyx Backend — User API

A RESTful API built with **Go**, **GoFiber**, **SQLC**, **PostgreSQL**, **Uber Zap**, and **go-playground/validator**.  
Users are stored with a `name` and `dob`; age is calculated dynamically on every fetch using Go's `time` package.

---

## Tech Stack

| Concern | Library |
|---|---|
| HTTP framework | [GoFiber v2](https://gofiber.io/) |
| Database | PostgreSQL via `lib/pq` |
| Query generation | [SQLC](https://sqlc.dev/) |
| Validation | [go-playground/validator v10](https://github.com/go-playground/validator) |
| Logging | [Uber Zap](https://github.com/uber-go/zap) |
| Config | [godotenv](https://github.com/joho/godotenv) |
| Containerisation | Docker + Docker Compose |

---

## Project Structure

```
.
├── cmd/server/main.go          # Entry point: wires all layers and starts Fiber
├── config/config.go            # Loads env vars; exposes DSN()
├── db/
│   ├── migrations/             # Raw SQL migration files
│   └── sqlc/                   # SQLC-generated Go code (db.go, models.go, queries.go)
├── internal/
│   ├── handler/user_handler.go # HTTP layer — parse, validate, call service
│   ├── repository/             # Data-access layer wrapping SQLC Queries
│   ├── service/user_service.go # Business logic, including CalculateAge()
│   ├── routes/routes.go        # Route registration
│   ├── middleware/middleware.go # RequestID, Logger, Recover
│   ├── models/user.go          # Request / response structs
│   └── logger/logger.go        # Uber Zap singleton initialiser
├── Dockerfile
├── docker-compose.yml
└── sqlc.yaml
```

---

## Quick Start (Docker — recommended)

```bash
# 1. Clone the repo
git clone https://github.com/udayagiri/ainyx-backend.git
cd ainyx-backend

# 2. Start Postgres + API together
docker compose up --build

# The API is now available at http://localhost:8080
```

The `docker-entrypoint-initdb.d` mechanism runs the migration automatically on
first Postgres start, so no manual `psql` step is needed.

---

## Local Setup (without Docker)

### Prerequisites

- Go 1.22+
- PostgreSQL 14+
- (Optional) [SQLC CLI](https://docs.sqlc.dev/en/latest/overview/install.html) if you want to regenerate queries

### Steps

```bash
# 1. Create the database
psql -U postgres -c "CREATE DATABASE ainyx_db;"
psql -U postgres -d ainyx_db -f db/migrations/001_create_users.sql

# 2. Copy and edit environment variables
cp .env.example .env
# Edit DB_PASSWORD (and anything else) in .env

# 3. Install Go dependencies
go mod download

# 4. Run the server
go run ./cmd/server
```

Server starts on `http://localhost:8080` (or `SERVER_PORT` from `.env`).

---

## API Reference

### Health Check

```
GET /health
→ 200  {"status":"ok"}
```

### Create User

```
POST /users
Content-Type: application/json

{ "name": "Alice", "dob": "1990-05-10" }

→ 201
{ "id": 1, "name": "Alice", "dob": "1990-05-10" }
```

### Get User by ID (includes age)

```
GET /users/1

→ 200
{ "id": 1, "name": "Alice", "dob": "1990-05-10", "age": 35 }
```

### Update User

```
PUT /users/1
Content-Type: application/json

{ "name": "Alice Updated", "dob": "1991-03-15" }

→ 200
{ "id": 1, "name": "Alice Updated", "dob": "1991-03-15" }
```

### Delete User

```
DELETE /users/1

→ 204 No Content
```

### List All Users (paginated)

```
GET /users?page=1&page_size=10

→ 200
{
  "data": [
    { "id": 1, "name": "Alice", "dob": "1990-05-10", "age": 35 }
  ],
  "total": 1,
  "page": 1,
  "page_size": 10,
  "total_pages": 1
}
```

---

## Running Tests

```bash
# Unit tests (no DB required — tests only pure Go logic)
go test ./internal/service/... -v
```

The test suite covers `CalculateAge` across normal birthdays, leap-year DOBs,
same-day edge cases, and birthday-not-yet-occurred scenarios.

---

## Regenerating SQLC Code

If you modify `db/sqlc/query.sql` or the migration:

```bash
# Install sqlc (once)
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Regenerate
sqlc generate
```

---

## Design Decisions

### Why is `age` not stored in the DB?
Age changes every day; storing it would require a background job to keep it
fresh. Calculating it at query time from `dob` using Go's `time` package is
correct, cheap, and always accurate.

### Age calculation algorithm (`CalculateAge`)
```
years = currentYear − birthYear
if birthday has not yet occurred this calendar year:
    years -= 1
```
The birthday-has-not-occurred check compares month and day by constructing a
`time.Time` for the birthday in the current year and checking `now.Before(...)`.
This correctly handles leap-year DOBs (Feb 29) on non-leap years.

### Layered architecture
```
Handler → Service → Repository → SQLC → PostgreSQL
```
Each layer depends only on the interface of the layer below it, making unit
testing straightforward without a real database.

### Pagination
`GET /users` accepts `?page` and `?page_size` (default 1 / 10, max 100).
The response envelope includes `total`, `total_pages`, `page`, and `page_size`
so clients can build pagination controls without extra requests.

### Middleware
- **RequestID** — injects `X-Request-ID` (preserves the caller's ID if present).
- **Logger** — logs method, path, status, and duration via Uber Zap after each request.
- **Recover** — catches panics and returns HTTP 500 instead of crashing the process.
