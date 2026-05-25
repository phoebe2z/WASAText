
# WASAText 

A WhatsApp-inspired instant messaging web application built as a university course project. Users can register, chat in real time, send images, manage contacts, and customize their profile — all in a clean, familiar interface.

---

## Tech Stack

| Layer | Technology |
|---|---|
| Frontend | Vue 3 + Vite |
| Backend | Go 1.24 |
| Database | SQLite (via `modernc.org/sqlite`) |
| Container | Docker + Docker Compose |
| Web Server | Nginx (frontend serving) |
| Router | `httprouter` |
| Logging | `logrus` |

---

## Features

- User registration and login (username-based, no password required)
- One-on-one conversations
- Send text messages and images
- Update username and profile photo
- View conversation history
- Static file serving for uploaded media

---

## Project Structure

```
WASAText/
├── cmd/
│   └── webapi/          # Go entrypoint (main.go)
├── service/             # Business logic & database layer
├── webui/               # Vue 3 frontend (Vite build)
│   └── nginx.conf       # Nginx config for production container
├── static/              # Uploaded files (avatars, images)
├── doc/                 # API specification (OpenAPI/YAML)
├── demo/                # Demo assets
├── Dockerfile.backend   # Multi-stage Go build → Debian slim
├── Dockerfile.frontend  # Multi-stage Node build → Nginx
├── docker-compose.yml   # Orchestrates both services
├── go.mod / go.sum
└── probe_db.go          # Database connectivity probe
```

---

## Getting Started

### Option 1 — Docker Compose (recommended)

Make sure you have [Docker](https://www.docker.com/) and Docker Compose installed.

```bash
git clone https://github.com/phoebe2z/WASAText.git
cd WASAText
docker-compose up --build
```

| Service | URL |
|---|---|
| Frontend | http://localhost:8080 |
| Backend API | http://localhost:3000 |

Data is persisted in a local `./data/` directory (SQLite database) and `./static/` (uploaded files).

---

### Option 2 — Run Locally

#### Backend

Requirements: Go 1.24+, CGO enabled (for SQLite)

```bash
go mod download
go build -o webapi ./cmd/webapi
./webapi
```

The server listens on port `3000` by default. The debug port is `4000`.

#### Frontend

Requirements: Node 18+, Yarn

```bash
cd webui
yarn install
yarn dev          # development server (hot reload)
yarn build-prod   # production build → dist/
```

---

## Configuration

The backend reads configuration via environment variables or a config file (using `ardanlabs/conf`):

| Variable | Default | Description |
|---|---|---|
| `CFG_DB_FILENAME` | `./decaf.db` | Path to SQLite database file |
| `CGO_ENABLED` | `1` | Required for SQLite |

---

## API

The REST API is defined in `doc/api.yaml` (OpenAPI format). Key endpoints include user registration, login, conversation management, and message/image sending.

The backend uses `julienschmidt/httprouter` for routing and `gorilla/handlers` for CORS and middleware support.

---

## Docker Details

Two separate Dockerfiles are used:

**`Dockerfile.backend`** — multi-stage build:
1. `golang:1.24` — compiles the `webapi` binary
2. `debian:bookworm-slim` — runs the binary with minimal footprint

**`Dockerfile.frontend`** — multi-stage build:
1. `node:18-alpine` — installs dependencies and runs `vite build --mode production`
2. `nginx:alpine` — serves the built `dist/` folder

Both containers share a `wasatext-network` bridge network defined in `docker-compose.yml`.

---

## License

[MIT](./LICENSE)
