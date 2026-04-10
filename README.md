# fraud-detection

Real-time fraud detection service built with Go, PostgreSQL, Redis, and Kafka.

## Prerequisites

- Go 1.26.1
- Docker & Docker Compose

## Setup

1. Clone the repository:

```bash
git clone https://github.com/Ahmet-Vt/fraud-detection.git
cd fraud-detection
```

2. Create your `.env` file from the example:

```bash
cp .env.example .env
```

3. Fill in your actual values in `.env` (database credentials, ports, etc.).

4. Start the infrastructure services:

```bash
docker compose up -d
```

5. Run the API in development mode (auto-rebuilds on file changes):

```bash
make daemon
```

The server starts on `http://localhost:8080`.

## Available Make Commands

| Command | Description |
|---|---|
| `make build` | Compile the binary to `./bin/api` |
| `make run` | Build and run |
| `make daemon` | Auto-rebuild and restart on file changes |

## API Endpoints

| Method | Path | Description |
|---|---|---|
| GET | `/ping` | Returns `{"message": "pong"}` |
| GET | `/health` | Checks Postgres connectivity |

## Project Structure

```
fraud-detection/
├── cmd/api/          Entry point
│   └── main.go
├── internal/
│   ├── config/       Environment loading and app configuration
│   │   └── config.go
│   └── database/     Database connection constructors
│       └── postgres.go
├── .env.example      Template for environment variables
├── docker-compose.yml
├── Makefile
├── go.mod
└── go.sum
```

## Infrastructure

All services are configured via environment variables in `.env` and run through Docker Compose:

- **PostgreSQL 16** — primary data store
- **Redis 7** — caching / session store
- **Kafka + Zookeeper** — event streaming
