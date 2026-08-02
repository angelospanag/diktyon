# API — Diktyon Go backend

Go service that supports multiple company registries via a provider abstraction, currently wired to Companies House (UK).

## Stack

| Concern | Package |
|---|---|
| HTTP framework | Huma v2 + chi v5 |
| Config | `caarlos0/env/v11` (struct tags on `internal/config/Config`) |
| Cache | `redis/go-redis/v9` + in-memory fallback (`sync.Map`) |
| Concurrency | `golang.org/x/sync/errgroup` |
| Rate limiting | `golang.org/x/time/rate` (token bucket) |
| Module path | `github.com/angelospanag/diktyon` |

## Directory structure

```
api/
├── cmd/api/main.go             Entry point: config → cache → providers → registry → router → server
├── internal/
│   ├── api/
│   │   ├── routes.go           Huma route registration; accepts *registry.Registry + cache
│   │   └── middleware/
│   │       └── logging.go      Structured request logging (slog)
│   ├── registry/
│   │   ├── provider.go         Provider interface + SearchResult + sentinel errors
│   │   └── registry.go         Registry (country-code → Provider map)
│   ├── providers/              One sub-package per jurisdiction
│   │   ├── uk/
│   │   │   ├── client.go       Companies House HTTP client (rate-limited, retrying, Basic auth)
│   │   │   ├── types.go        CH API response structs
│   │   │   ├── builder.go      ForCompany / ForOfficer — UK-specific fan-out graph logic
│   │   │   └── provider.go     registry.Provider implementation
│   │   └── gr/
│   │       ├── client.go       GEMI Open Data HTTP client
│   │       ├── types.go        GEMI API response structs
│   │       └── provider.go     registry.Provider implementation
│   ├── cache/
│   │   ├── cache.go            Cache interface + ErrCacheMiss sentinel
│   │   ├── redis.go            Redis implementation
│   │   ├── memory.go           In-memory implementation (dev / Redis-unavailable fallback)
│   │   └── noop.go             No-op implementation (tests)
│   ├── config/
│   │   └── config.go           Env-var config struct
│   └── graph/
│       └── types.go            Node / Edge / Response types + Collect helper
```

## Provider pattern

New country support = new package implementing `registry.Provider`:

```go
type Provider interface {
    SearchCompanies(ctx, query) ([]SearchResult, error)
    CompanyGraph(ctx, companyID, depth) (*graph.Response, error)
    OfficerGraph(ctx, officerID) (*graph.Response, error)  // return ErrUnsupported if not available
}
```

Register it in `cmd/api/main.go`:
```go
reg.Register("xx", xx.New(xx.NewClient(...)))
```

All routes accept `?country=` (default `uk`). Cache keys are prefixed with country.

## Environment variables

| Variable | Required | Default | Notes |
|---|---|---|---|
| `COMPANIES_HOUSE_API_KEY` | yes | — | Basic auth username; password is always empty |
| `PORT` | no | `8080` | |
| `REDIS_URL` | no | — | If unset, uses in-memory cache |
| `CACHE_TTL` | no | `24h` | Default TTL for cached responses |
| `LOG_LEVEL` | no | `INFO` | `DEBUG`, `INFO`, `WARN`, `ERROR` |
| `RATE_LIMIT_RPS` | no | `2` | Sustained CH API calls per second |
| `RATE_LIMIT_BURST` | no | `10` | Token-bucket burst |
| `GEMI_API_KEY` | no | — | Set to any non-empty value to enable the experimental `gr` (GEMI) provider |

## Cache key conventions

```
search:v1:{country}:{query}                        # 1-hour TTL (search results are time-sensitive)
graph:v5:{country}:company:{number}:depth:{1|2}    # default TTL (24h)
graph:v5:{country}:officer:{id}                    # default TTL (24h)
```

Bump the `v1` segment when the cached shape changes — this avoids serving stale structure from Redis.

## Graph builder

`internal/providers/uk/builder.go` — `Builder.ForCompany` and `Builder.ForOfficer`.

- **ForCompany(depth=1)**: fans out three concurrent CH calls (profile, officers, PSCs) via `errgroup`, deduplicates into `map[string]Node` + `map[string]Edge`, returns a flat node/edge list.
- **ForCompany(depth=2)**: additionally fetches appointments for the first 5 active officers (capped to limit rate-limit exposure).
- **ForOfficer**: fetches all current appointments for one officer.

Node ID scheme:
- `company:{company_number}` — stable across all graph expansions; corporate PSCs with a UK number reuse this prefix so they merge automatically.
- `officer:{officer_id}` — extracted from the `/officers/{id}/appointments` URL in the CH response.
- `psc:{sanitized_name}` — for non-UK or unregistered PSCs that have no company number.

Polaroid card tilt is derived client-side from a hash of the node ID (`ui/lib/graph.ts`'s `tiltAngle`).

## Dev

```bash
cd api
cp .env.example .env   # fill in COMPANIES_HOUSE_API_KEY
go run ./cmd/api    # starts on :8080
```

## Testing

```bash
cd api
go test ./...
```

CI runs tests with a live Redis container (`REDIS_URL=redis://localhost:6379`).

## Adding a new endpoint

1. Add input/output structs in `internal/api/routes.go`.
2. Register with `huma.Register(api, huma.Operation{...}, handler)`.
3. Call through `reg.Get(country)` to get the right provider.

## Adding a new country

1. Create `internal/providers/{cc}/` with `client.go`, `types.go`, and `provider.go` implementing `registry.Provider`.
2. Register in `cmd/api/main.go`: `reg.Register("xx", xx.New(xx.NewClient(...)))`.
3. Add any required API key env vars to `internal/config/config.go`.

Huma generates OpenAPI docs automatically at `/docs` (Swagger UI).
