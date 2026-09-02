# Task Management API

REST API untuk manajemen task multi-user, dibangun dengan Go menggunakan **Clean Architecture**.

## Tech Stack

- **Language:** Go 1.25
- **Framework:** Gin Gonic
- **Database:** PostgreSQL
- **Authentication:** JWT
- **Testing:** Go test dengan testify

## Struktur Proyek

```
task-api/
├── cmd/api/               # Entry point aplikasi
├── internal/
│   ├── domain/            # Entitas & interface (kontrak)
│   ├── usecase/           # Business logic & validasi
│   ├── repository/         # Implementasi database (PostgreSQL)
│   ├── delivery/           # HTTP handlers & middleware
│   │   └── http/
│   │       ├── task_handler.go
│   │       ├── user_handler.go
│   │       └── middleware/
│   └── pkg/response/      # Standardized response & error handling
├── docker-compose.yml
├── Dockerfile
└── init.sql               # Schema database
```

## Arsitektur

Proyek ini menggunakan **Clean Architecture** dengan 4 layer:

| Layer | Responsibility |
|-------|----------------|
| **Domain** | Entitas bisnis dan interface (kontrak repository) |
| **Usecase** | Business logic, validasi, dan idempotency |
| **Repository** | Implementasi interaksi dengan PostgreSQL |
| **Delivery** | HTTP Handlers dan Middleware (Gin) |

## Fitur

### 1. Idempotency
Menggunakan tabel `idempotency_keys` dengan constraint `UNIQUE`. Request `POST /tasks` dengan header `Idempotency-Key` yang sama akan mengembalikan response yang sama tanpa duplikasi data.

### 2. Structured Error Handling
- `PanicRecovery` middleware menangkap panic dan mengembalikan JSON 500
- Error client (4xx) dan server (5xx) distandarisasi menggunakan package `response`

### 3. Database Transaction
Endpoint `/tasks/:id/assign` dibungkus dalam transaction. Jika `UPDATE` atau `INSERT` gagal, maka di-rollback.

### 4. Logging & Observability
Middleware log mencatat `request_id`, method, path, status, dan latency dalam format JSON ke stdout.

### 5. Unit Testing
Mock repository digunakan untuk testing business logic, termasuk simulasi concurrent request.

## Endpoint API

| Method | Endpoint | Deskripsi |
|--------|----------|-----------|
| `POST` | `/register` | Registrasi user baru |
| `POST` | `/login` | Login, returns JWT token |
| `POST` | `/tasks` | Create task (header: `Idempotency-Key`) |
| `GET` | `/tasks` | List tasks (query: `status`, `title`, `page`, `limit`) |
| `GET` | `/tasks/:id` | Get detail task |
| `PUT` | `/tasks/:id` | Update task |
| `DELETE` | `/tasks/:id` | Delete task |
| `POST` | `/tasks/:id/assign` | Assign task ke user lain (transaction) |

## Prerequisites

- Go 1.25+
- Docker & Docker Compose

## Cara Menjalankan

```bash
# 1. Clone repository
git clone <repo-url>
cd task-management-api

# 2. Jalankan dengan Docker Compose
docker-compose up -d --build

# 3. API tersedia di
http://localhost:8080
```

## Environment Variables

| Variable | Default | Deskripsi |
|----------|---------|-----------|
| `DB_HOST` | `localhost` | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USER` | `postgres` | Database user |
| `DB_PASSWORD` | `postgres` | Database password |
| `DB_NAME` | `taskdb` | Database name |
| `JWT_SECRET` | `your-secret-key` | Secret key untuk JWT |
| `PORT` | `8080` | Port aplikasi |
