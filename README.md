## PROJECTS PURPOSE AND SCOPE

This repository is a **fraud-detection** demo stack: a **Go (Gin)** backend stores users and transactions in **PostgreSQL**, publishes new transactions to **RabbitMQ**, and runs **asynchronous fraud evaluation** (rules such as amount, location, and velocity). Results are written back to the database, cached in **Redis** where relevant, and streamed to a **React (Vite + TypeScript)** dashboard over **Server-Sent Events (SSE)**. Optional **bash** scripts help generate load or single transactions for testing. An **MCP (Model Context Protocol)** server over **stdio** exposes read-only tools for agents (recent frauds, user status).

---

## SYSTEM ARCHITECTURE AND COMPONENTS EXPLANATION

Infrastructure runs in **Docker** (`[docker-compose.yml](docker-compose.yml)`): **PostgreSQL**, **Redis**, **RabbitMQ**, `**api`** (`[backend/Dockerfile](backend/Dockerfile)`), `**web`** (`[fraud-dashboard/Dockerfile](fraud-dashboard/Dockerfile)`). The backend uses **handler → service → repository**; RabbitMQ consumers and optional stdio MCP run **in the same `api` process** as Gin. Local dashboard dev uses **Vite** (`[fraud-dashboard/vite.config.ts](fraud-dashboard/vite.config.ts)`); the `**web`** image serves the built SPA and proxies `**/api`** to `**api:8080**`.

### Architecture diagram

`**[docs/architecture.png](docs/architecture.png)**` 

System architecture

---

## REQUIREMENTS (what to install)


| Goal                                                    | Install                                                                                                                                                                                                                           |
| ------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Full stack in Docker only**                           | **Docker** + `**docker compose`**. Nothing else is required. Images for `**api`** and `**web**` are built automatically when you run `**docker compose up --build**` — you do **not** run `docker build` by hand for the app.     |
| **Backend on the host** (`go run ./cmd/api`)            | **Go 1.26+** (see `[backend/go.mod](backend/go.mod)`), `**backend/.env`**, and Postgres / Redis / RabbitMQ reachable (e.g. the same Compose stack without the `api` container, or full Compose with a different `API_HOST_PORT`). |
| **Dashboard on the host** (`npm run dev`)               | **Node.js** + **npm** in `[fraud-dashboard/](fraud-dashboard/)` (`npm ci` then `npm run dev`).                                                                                                                                    |
| **Load / manual bash scripts**                          | `**bash`**, `**curl`**, `**jq**` (`[backend/scripts/](backend/scripts/)`). On Windows, use **Git Bash** or **WSL** if `bash` is missing.                                                                                          |
| **MCP Inspector (example in [MCP USAGE](#mcp-usage))**  | **Node** / `**npx`** for `@modelcontextprotocol/inspector`.                                                                                                                                                                       |
| **Regenerate Swagger after changing Go route comments** | `**swag`** CLI (`go install github.com/swaggo/swag/cmd/swag@latest`); committed docs live under `[backend/docs/](backend/docs/)`. Optional for normal use.                                                                        |
| **Optional dev auto-rebuild**                           | `[backend/Makefile](backend/Makefile)` target `**daemon`** uses **CompileDaemon** (`go install github.com/githubnemo/CompileDaemon@latest`).                                                                                      |


---

## PROJECT SETUP STEPS

### 1. Clone and enter the repo

```bash
git clone <repository-url> fraud-detection
cd fraud-detection
```

### 2. Environment: `backend/.env`

Create the file from the example, then edit values (passwords, ports, etc.):

```bash
cp backend/.env.example backend/.env
```

- **Docker-only:** you still need `**backend/.env`** — the `**api`** service loads it via `env_file` in `[docker-compose.yml](docker-compose.yml)`. The app does **not** read the repo-root `.env` by itself.
- **Root `.env`:** only used by Compose for `**${POSTGRES_PORT}`**, `**${API_HOST_PORT}`**, etc. Easiest: same file as `backend/.env` (symlink on Linux/macOS, or copy on Windows).
- `**go run` on the host:** keep `**POSTGRES_HOST`**, `**REDIS_HOST`**, `**RABBITMQ_HOST**` as `**localhost**`. Inside the `**api**` container, Compose overrides those to `**postgres**`, `**redis**`, `**rabbitmq**`.

**Symlink (Linux / macOS) — one source of truth for Compose + `backend/.env`:**

```bash
ln -sf backend/.env .env
```

**Windows (no `ln`):** copy `backend\.env` to `.env` at the repo root when you change settings.

**Alternative — no root `.env`:**

```bash
docker compose --env-file backend/.env up -d --build
```

### 3. Docker Compose

From the **repository root** (after step 2 (defining the .env file)):

```bash
docker compose up -d --build
```

- `**--build**` builds or rebuilds the `**api**` and `**web**` images when needed — you do **not** run `docker build` separately for these services.
- **Services:** `postgres`, `redis`, `rabbitmq`, `api` (HTTP + Rabbit consumers in one process), `web` (static UI; Nginx proxies `**/api`** → `**api:8080`** inside the network).

**After you change application source** (`[backend/](backend/)` Go code or `[fraud-dashboard/](fraud-dashboard/)` UI): Compose does **not** bind-mount the repo into `**api**` or `**web**`; each image bakes in a built binary or static assets. A container that is already running keeps the old build until the image is rebuilt and the service recreated. Examples:

```bash
docker compose up -d --build api   # backend only
docker compose up -d --build web   # dashboard only
docker compose up -d --build       # both app images
```

If you run the API on the host (`**go run ./cmd/api**`), you skip this rebuild step for backend changes.

**Check that containers are up:**

```bash
docker compose ps
```

**Check the API** (default host port **8080**; use your `**API_HOST_PORT`** from `backend/.env` if different):

```bash
curl -sS http://localhost:8080/health
```

**Open the UI** (default `**WEB_HOST_PORT` 8088**):

```text
http://localhost:8088/
```

- Do not use the same **host** port twice (e.g. `**go run`** on **8080** while `**API_HOST_PORT=8080`** maps Docker there).
- Use `**docker compose`** (or legacy `**docker-compose**` with the same `[docker-compose.yml](docker-compose.yml)`).

**Stop the stack:**

```bash
docker compose down
```

### 4. Database migrations

No manual step: migrations run when the API process starts (`[backend/cmd/api/main.go](backend/cmd/api/main.go)`), in Docker or on the host.

### 5. Run the API on the host (instead of the `api` container)

```bash
cd backend
go run ./cmd/api
```

**Or** build a binary and run it:

```bash
cd backend
go build -o bin/api ./cmd/api
./bin/api
```

**Optional** (`[backend/Makefile](backend/Makefile)`): `make build`, `make run`, or `make daemon` (auto-rebuild while developing).

Listen address: env `**HTTP_LISTEN`** (default `**:8080`**, see `[backend/internal/config/config.go](backend/internal/config/config.go)`).

### 6. Run the dashboard on the host (Vite dev server)

```bash
cd fraud-dashboard
npm ci
npm run dev
```

Then open the URL Vite prints (often `**http://localhost:5173**`). The dev server proxies `**/api/***` to `**http://localhost:8080**` (see `[fraud-dashboard/vite.config.ts](fraud-dashboard/vite.config.ts)` and `[fraud-dashboard/src/api/endpoints.ts](fraud-dashboard/src/api/endpoints.ts)`).

---

## USAGE GUIDE

- **Docker UI:** after `docker compose up`, open `**http://localhost:8088/`** .
- **Local dev UI:** `cd fraud-dashboard && npm run dev` → usually `**http://localhost:5173`**; Vite proxies `**/api`** to `**http://localhost:8080**` (`[vite.config.ts](fraud-dashboard/vite.config.ts)`).
- **Swagger:** `http://localhost:<API_HOST_PORT>/swagger/index.html` (default host port **8080**).
- **Live transactions** page consumes **SSE** at `**/stream/transactions`** (proxied from the dashboard in dev).
- **Users** pages list users and show per-user transaction charts/tables with optional date filters.

---

## API USAGE

High-level REST surface (see Swagger for schemas and status codes):


| Area         | Examples                                                                                                |
| ------------ | ------------------------------------------------------------------------------------------------------- |
| Users        | `POST /users`, `GET /users`, `GET /users/:id`                                                           |
| Transactions | `POST /transactions`, `GET /transactions`, `GET /transactions/:id`, `DELETE /transactions/:id`          |
| By user      | `GET /transactions/user/:id`, `GET /transactions/user/:id/detail`, `GET /transactions/user/:id/summary` |
| System       | `GET /health`                                                                                           |


Implementation entry points: `[backend/internal/user/handler.go](backend/internal/user/handler.go)`, `[backend/internal/transaction/handler.go](backend/internal/transaction/handler.go)`. Prefer **Swagger** for request/response models and query parameters.

---

## MCP USAGE

Behavior matches `[backend/cmd/api/main.go](backend/cmd/api/main.go)` and `[backend/internal/config/config.go](backend/internal/config/config.go)`:

1. `**config.Load()`** reads `**backend/.env`** (via `godotenv`; working directory should be `**backend/**`) then process environment. `**HTTP_LISTEN**` defaults to `**:8080**`. `**MCP_STDIO_ENABLED**` defaults to **false**.
2. The process **always** opens Postgres/Redis/Rabbit, runs migrations, starts **Rabbit fraud consumers**, builds **Gin**, and calls `**router.Run(HTTP_LISTEN)`** — that is the HTTP API.
3. If `**MCP_STDIO_ENABLED=true`**, it **also** starts `**mcp.RunStdio`** in a **goroutine**: MCP JSON-RPC on **stdin/stdout** (same OS process as Gin). MCP does not replace HTTP; both run until the process exits.
4. **Docker `api` service** (`[docker-compose.yml](docker-compose.yml)`): `**environment`** sets `**HTTP_LISTEN=:8080`** and `**MCP_STDIO_ENABLED=false**`, so the container only speaks HTTP on **8080** inside the network. `**ports`** is `**${API_HOST_PORT:-8080}:8080`**: that is **host port → container port 8080**, not the Go variable name; changing `**API_HOST_PORT`** does not change `**HTTP_LISTEN`** inside the app.
5. **Host `go build` + MCP Inspector** while the `**api`** container already uses host **8080**: the second process needs `**HTTP_LISTEN`** on **some other free TCP port**. If `**bind: address already in use`**, that port is busy — pick a different one, or stop what holds it (often a stuck `**bin/api`**), e.g. `**lsof -iTCP:8090 -sTCP:LISTEN**` (change **8090** to the port you tried).

Tools (`[backend/internal/mcp/server.go](backend/internal/mcp/server.go)`): `**get_recent_frauds`** (`from` / optional `to` in `**DD.MM.YYYY`**, optional `**limit**`); `**check_user_status**` (`user_id` UUID).

Example (Compose DB/Rabbit/Redis on localhost ports from `**backend/.env**`):

To connect to MCP Inspector with the commands below, the TCP port you set in `**HTTP_LISTEN**` (the port number after the colon, e.g. **8090**) **must be free on your host**—nothing else may be listening on that port. If it is already in use, choose another free port or stop the process that holds it.

```bash
cd fraud-detection/backend   
go build -o bin/api ./cmd/api/
HTTP_LISTEN=:8090 MCP_STDIO_ENABLED=true npx @modelcontextprotocol/inspector ./bin/api
```

(Replace **8090** with any **free** host port; use the same value in `**HTTP_LISTEN`** and when opening `**http://localhost:<port>/…`** for this binary’s HTTP API.)

- **Container API:** `http://localhost:<API_HOST_PORT>/…` (default host port **8080**).  
- **Rabbit:** this second process still starts `**RABBITMQ_CONSUMER_COUNT`** consumers unless you set that env to **0** for this run.

---

## SCRIPTS USAGE AND PARAMETERS

### `backend/scripts/auto-test.sh`

Sends synthetic `POST /transactions` traffic against a running API.


| Option                 | Meaning                                 | Default |
| ---------------------- | --------------------------------------- | ------- |
| `--duration=<seconds>` | How long the script runs                | `30`    |
| `--rate=<rps>`         | Requests per second                     | `5`     |
| `--anomaly=<0-100>`    | Probability of “anomaly-style” requests | `30`    |
| `-h`, `--help`         | Help                                    |         |


Environment: `**BASE_URL`** (default `http://localhost:8080`).

Example:

```bash
cd backend
./scripts/auto-test.sh --duration=20 --rate=5 --anomaly=50
```

### `backend/scripts/manual-input.sh`

Should have user before or it will return user not found error!

Creates **one** transaction via `POST /transactions`.

```text
./scripts/manual-input.sh <user_id> <amount> <location>
Example: ./scripts/manual-input.sh 1a24123e-c7dd-4817-895d-e7f4260d06ca 3000 istanbul
```

Requires `**curl**` and `**jq**` defaults to `http://localhost:8080`.

---

## MCP TEST DOCUMENTED

1. Compose up (infra + `**api**` if you want the container API on host **8080**).
2. From `**backend/`**, run the two commands under [MCP USAGE](#mcp-usage) (`npx` spawns `**./bin/api`** with the exported env).
3. Inspector **Connect** → tools `**get_recent_frauds`**, `**check_user_status`**.
4. Run `**get_recent_frauds**` with e.g. `from` `**01.05.2026**`, leave `**to**` empty for “until now”, or set `**02.05.2026**` for an end-of-day upper bound; expect JSON array of fraud transactions or a clear tool error string.
5. Run `**check_user_status**` with a valid `**user_id**` UUID; expect JSON with user fields including `**risk_rate**`.
6. Optional automated parse tests: `cd backend && go test ./internal/mcp/ -count=1`.

