# Order Pack Calculator

API: pack sizes in Postgres, **min packs** for exact order (change-making).

**1. Run** — From repo root (`go.mod`). **Docker stack:** the **`Makefile`** wraps Compose — run **`make up`** to build & start app + Postgres, **`make down`** to stop, **`make logs`** for app logs. Needs **Docker + Compose v2**. **Host ports free:** **8080** (API; override in `app.ports` if needed), **5432** (Postgres for IDE/psql — `postgres.ports` in compose; use e.g. `5433:5432` if 5432 is taken). **No Docker:** Go 1.24+, Postgres 16, `.env` from `.env.example`, `go run ./cmd/calculator`.

**2. Health** — `curl -s http://127.0.0.1:8080/live` · `/ready` → `"status":"ok"`. `/ready` needs DB.

**3. Edge case (500 000 items, packs 23/31/53)**

```bash
curl -s -X PUT http://127.0.0.1:8080/api/v1/pack-sizes -H 'Content-Type: application/json' -d '{"sizes":[23,31,53]}'
curl -s -X POST http://127.0.0.1:8080/api/v1/calculate -H 'Content-Type: application/json' -d '{"items":500000}'
```

→ `packs`: 23×2, 31×7, 53×9429, `"message":"ok"`.

**4. Architecture**

- **api** (`internal/api/http`) — Gin server, routes, controllers; validation; errors as `{"error":"..."}`.
- **app** (`internal/app`) — env config, dependency wiring, lifecycle.
- **domain** (`internal/domain`) — packing math, errors; core: `minpacks.go`.
- **repository** (`internal/repository/pg`) — `pgxpool` without extra wrapping, migrations, pack sizes persistence.
- **usecase** (`internal/usecase/calculator`) — orchestrates store + domain, business validation.

Flow: `cmd/calculator` → `app` → HTTP → **usecase** → **repository** / **domain**.

**5. Tests** — `make test` · `make test-short` · `make test-integration` (`-tags=integration`, Docker). **Mocks:** `make mocks` (`.mockery.yaml`).

**6. Questions you may have** (design rationale)

- **Layers (api / usecase / domain / repo)** — keeps HTTP, orchestration, pure math, and SQL apart so each is testable in isolation and the packing algorithm stays a small, reviewable core.
- **`PUT /api/v1/pack-sizes`** — replaces the **entire** saved size list in one call; PUT fits “full resource replace” semantics (idempotent for the same body).
- **Validation in both use case and domain** — use case rejects bad input early (e.g. `items` before hitting DB) and returns typed errors for the API; domain still guards its own contract so `MinPacks` stays safe when called from tests or future callers (gRPC, CLI and so on).
- **One error shape: `{"error":"..."}`** — same JSON for calculator and pack-size errors; success calculate responses still use `packs` + `message`.
- **Integration tests behind `-tags=integration` + Testcontainers** — default `go test ./...` stays fast and needs no Docker; full repo + Postgres runs only when you opt in.
- **Mockery mocks next to consumers** — `UseCase` interface lives in the HTTP package (and `PackStore` at the use case) so tests don’t import production wiring just to stub a port.
- **Alpine runtime (not distroless)** — small image plus `wget` so Compose can `HEALTHCHECK` against `/ready` (DB + app); distroless has no shell/curl for that probe.
- **Postgres `5432` published** — optional local access from IDE/psql; change the host port in compose if it clashes with another Postgres.
- **`godotenv.Load(".env")`** — In a real production service you’d usually **omit** this and inject config only via the environment (K8s, systemd, etc.). It’s left here for **local dev**: copy `.env.example` → `.env` in the repo root when running `go run ./cmd/calculator`. In **Docker**, the app’s working directory normally has **no** `.env` file, so the load is a no-op and variables come from `docker-compose` / the runtime.