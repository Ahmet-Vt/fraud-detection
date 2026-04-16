# fraud-detection

Fraud detection backend written in Go with Gin and PostgreSQL.

The current codebase includes:

- API bootstrap under `cmd/api`
- PostgreSQL connection and startup migrations
- `user` module with layered `handler -> service -> repository` flow
- safe/public vs internal error separation via `internal/secureerr`

Redis and RabbitMQ are provisioned in `docker-compose.yml`. PostgreSQL is fully wired into the application, and RabbitMQ is used both during startup and in runtime request handling.

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

RabbitMQ management UI is available at `http://localhost:15672` with the credentials configured in `.env`.

## Swagger

After starting the API, open:

`http://localhost:8080/swagger/index.html`

This page shows all documented endpoints, request bodies, response models, and status codes.

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
4. Connects to RabbitMQ.
5. Wires the `user` and `transaction` modules.
6. Starts the Gin HTTP server.

RabbitMQ connection settings are read from `.env` using `RABBITMQ_HOST`, `RABBITMQ_PORT`, `RABBITMQ_USER`, and `RABBITMQ_PASSWORD`.

## API Endpoints

| Method | Path | Description |
| --- | --- | --- |
| `GET` | `/health` | Checks PostgreSQL and RabbitMQ connectivity |
| `POST` | `/users` | Creates a user |
| `GET` | `/users` | Returns all users |
| `GET` | `/users/:id` | Returns a user by UUID |
| `POST` | `/transactions` | Creates a transaction |
| `GET` | `/transactions` | Returns all transactions |
| `GET` | `/transactions/:id` | Returns a transaction by UUID |
| `GET` | `/transactions/user/:id` | Returns transactions for a user |
| `DELETE` | `/transactions/:id` | Deletes a transaction |

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
- RabbitMQ 4 with the management plugin

The running API depends on both PostgreSQL and RabbitMQ.
