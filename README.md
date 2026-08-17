# Shortlink API

A production-shaped Go short link service with JWT authentication, MySQL storage,
redirect tracking, access statistics, and a basic dashboard API.

## Stack

- Go 1.23
- `net/http` with `github.com/go-chi/chi/v5`
- MySQL 8.4 via `database/sql` and `github.com/go-sql-driver/mysql`
- `github.com/golang-jwt/jwt/v5`
- `golang.org/x/crypto/bcrypt`
- Docker Compose

## Directory Layout

```text
cmd/server/                  service entrypoint
internal/user/               user registration/login
internal/shortlink/          short link create/list/delete
internal/access/             access stats and dashboard
internal/redirect/           public short code redirect
migrations/                  SQL migrations applied on startup
pkg/                         shared config, database, JWT, middleware, utilities
```

Each `internal` module is split into `handler`, `service`, `repository`, and
`model` packages.

## Run

```bash
docker compose up -d --build
```

The Go service listens on host port `18094` and container port `8080`. The app
waits for MySQL readiness and applies every `.sql` file under `migrations/` in
lexical order.

Default environment values are already set in `docker-compose.yml`. Copy
`.env.example` to `.env` if you need to change them.

## Default Demo Account

The startup seed creates:

```text
username: demo
password: demo123456
```

It also creates two demo links: `demo-home` and `demo-expired`.

## API

Public endpoints:

```text
POST /api/v1/auth/register
POST /api/v1/auth/login
GET  /r/{short_code}
GET  /healthz
```

Authenticated endpoints require `Authorization: Bearer <token>`:

```text
GET    /api/v1/links
POST   /api/v1/links
GET    /api/v1/links/{id}
DELETE /api/v1/links/{id}
GET    /api/v1/stats/links/{id}
GET    /api/v1/dashboard
```

Create link request:

```json
{
  "original_url": "https://example.com/path",
  "custom_alias": "optional-alias",
  "expires_at": "2026-08-20T00:00:00Z"
}
```

`custom_alias` and `expires_at` are optional. When `custom_alias` is omitted, a
random six-character code is generated. Alias collisions return `409 Conflict`.

## Stop And Cleanup

```bash
docker compose down -v
```
