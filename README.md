# fraud-detection

Fraud detection backend written in Go with Gin and PostgreSQL.

The current codebase includes:

- API bootstrap under `cmd/api`
- PostgreSQL connection and startup migrations
- `user` module with layered `handler -> service -> repository` flow
- safe/public vs internal error separation via `internal/secureerr`

Redis and Kafka are present in `docker-compose.yml` for future work, but they are not wired into the application flow yet.

## Prerequisites

- Go `1.26.1`
- Docker and Docker Compose

## Setup

1. Clone the repository.

```bash
git clone https://github.com/Ahmet-Vt/fraud-detection.git
cd fraud-detection
```

2. Create the environment file.

```bash
cp .env.example .env
```

3. Start local infrastructure.

```bash
docker compose up -d
```

4. Run the API.

```bash
make run
```

For development with automatic rebuilds:

```bash
make daemon
```

The API listens on `http://localhost:8080`.

## Available Make Commands

| Command | Description |
| --- | --- |
| `make build` | Build the API binary into `./bin/api` |
| `make run` | Build and run the API |
| `make daemon` | Run with auto rebuild on file changes |

## Startup Flow

On startup the API:

1. Loads configuration from environment variables.
2. Connects to PostgreSQL.
3. Registers and runs migrations.
4. Wires the `user` repository, service, and handler.
5. Starts the Gin HTTP server.

## API Endpoints

| Method | Path | Description |
| --- | --- | --- |
| `GET` | `/ping` | Returns `{"message":"pong"}` |
| `GET` | `/health` | Checks PostgreSQL connectivity |
| `POST` | `/users` | Creates a user |
| `GET` | `/users` | Returns all users |
| `GET` | `/users/:id` | Returns a user by UUID |

### Create User Request

```json
{
  "name": "Ahmet",
  "surname": "Vt",
  "email": "ahmet@example.com",
  "status": "active"
}
```

Valid `status` values:

- `active`
- `suspended`
- `deleted`

### Error Response Format

Application errors are exposed with a safe response shape:

```json
{
  "code": "USER_NOT_FOUND",
  "error": "user not found"
}
```

Internal details are kept inside the application and are only intended for logs.

## Current Layering

### `internal/user/handler.go`

- Owns HTTP concerns
- Binds and validates requests
- Parses route parameters
- Maps domain errors to safe API errors

### `internal/user/service.go`

- Owns business flow
- Calls repositories
- Wraps errors with context
- Does not know about HTTP status codes

### `internal/user/repository.go`

- Owns database access
- Translates low-level database failures into meaningful sentinel errors

### `internal/secureerr/error.go`

- Separates public-safe error messages from internal causes
- Supports structured logging without leaking internals to clients

## Project Structure

```text
fraud-detection/
├── cmd/api/
│   └── main.go
├── internal/
│   ├── config/
│   ├── database/
│   ├── domain/
│   ├── secureerr/
│   └── user/
├── migrations/
├── .env.example
├── docker-compose.yml
├── Makefile
├── go.mod
└── go.sum
```

## Infrastructure

Docker Compose currently provisions:

- PostgreSQL 16
- Redis 7
- Kafka
- Zookeeper

At the moment, only PostgreSQL is used by the running API.
