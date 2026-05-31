# LoomBloom

LoomBloom is a Go + PostgreSQL backend for loom owners to manage daily factory operations: machines, workers, stock, spare parts, raw materials, payments, production entries, and subscriptions.

## Tech Stack

- Go 1.22+
- PostgreSQL 15+
- Standard Go HTTP server
- `pgx` PostgreSQL driver
- No Docker

## Project Layout

```text
cmd/server/              application entrypoint
internal/config/         environment configuration
internal/httpapi/        HTTP routes and handlers
internal/store/          PostgreSQL access layer
migrations/              SQL migrations
```

## Local Development

1. Create a local PostgreSQL database.

```powershell
createdb loombloom
```

2. Create `.env` from the example and update `DATABASE_URL`.

```powershell
Copy-Item .env.example .env
$env:DATABASE_URL="postgres://loombloom:loombloom@localhost:5432/loombloom?sslmode=disable"
```

3. Run migrations.

```powershell
make migrate-up
```

If `make` is not installed on Windows, run:

```powershell
psql "$env:DATABASE_URL" -f migrations/001_init.up.sql
```

4. Start the API.

```powershell
go run ./cmd/server
```

The app listens on `http://localhost:8080` by default. Open `/` for the web dashboard and `/api/v1/modules` for the API module list.

Optional development seed data:

```powershell
psql "$env:DATABASE_URL" -f migrations/002_seed.dev.sql
```

## Production Setup Without Docker

1. Install Go and PostgreSQL on the server.
2. Create a dedicated PostgreSQL user and database.
3. Set environment variables in your process manager:

```text
APP_ENV=production
HTTP_ADDR=:8080
DATABASE_URL=postgres://USER:PASSWORD@HOST:5432/loombloom?sslmode=require
READ_TIMEOUT=5s
WRITE_TIMEOUT=10s
SHUTDOWN_TIMEOUT=10s
```

4. Build the binary.

```powershell
go build -o bin/loombloom ./cmd/server
```

5. Run migrations with the production `DATABASE_URL`.

```powershell
psql "$env:DATABASE_URL" -f migrations/001_init.up.sql
```

6. Run the binary with a service manager such as systemd, Windows Service, NSSM, or PM2.

## API Quick Check

```powershell
curl http://localhost:8080/healthz
curl http://localhost:8080/readyz
curl http://localhost:8080/api/v1/modules
```

Create an organization:

```powershell
curl -X POST http://localhost:8080/api/v1/organizations `
  -H "Content-Type: application/json" `
  -d "{\"name\":\"Shree Looms\",\"owner_name\":\"Mehul\",\"phone\":\"9999999999\"}"
```

## Core API Routes

- `GET /healthz`
- `GET /readyz`
- `GET /api/v1/modules`
- `GET /api/v1/organizations`
- `POST /api/v1/organizations`
- `GET /api/v1/organizations/{organizationID}/machines`
- `POST /api/v1/organizations/{organizationID}/machines`
- `GET /api/v1/organizations/{organizationID}/workers`
- `POST /api/v1/organizations/{organizationID}/workers`
- `GET /api/v1/organizations/{organizationID}/stock`
- `POST /api/v1/organizations/{organizationID}/stock`
- `GET /api/v1/organizations/{organizationID}/production`
- `POST /api/v1/organizations/{organizationID}/production`

## Next Production Steps

- Add authentication and roles for owner, manager, and staff users.
- Add full CRUD for spare parts, raw materials, and payments.
- Add reports for daily, weekly, monthly, and yearly production.
- Add database migration tooling once migration history grows.
- Add frontend/mobile app after the API contracts settle.
