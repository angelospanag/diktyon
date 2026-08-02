# Diktyon — API

Go backend that proxies multiple company registries and builds graph responses consumed by the UI.

- [Diktyon — API](#diktyon--api)
  - [Endpoints](#endpoints)
  - [Running locally](#running-locally)
  - [Configuration](#configuration)
  - [Provider architecture](#provider-architecture)
  - [Adding a new jurisdiction](#adding-a-new-jurisdiction)
  - [Testing](#testing)

## Endpoints

OpenAPI docs available at `http://localhost:8080/docs` when running locally.

All endpoints accept an optional `?country=` parameter (default: `uk`).

| Method | Endpoint                              | Description                                            |
| ------ | ------------------------------------- | ------------------------------------------------------ |
| GET    | `/api/search?q=&country=`             | Search companies by name (across supported registries) |
| GET    | `/api/company/{number}/graph?depth=1` | Company graph (depth 1 or 2)                           |
| GET    | `/api/officer/{id}/graph`             | All current appointments for an officer                |

## Running locally

```bash
cp .env.example .env
# edit .env — set COMPANIES_HOUSE_API_KEY at minimum
go run ./cmd/api
# API → http://localhost:8080
# Docs → http://localhost:8080/docs
```

Or via mise:

```bash
mise run dev   # from api/
# or from the repo root: mise -C api run dev
```

## Configuration

| Variable                  | Required | Default | Notes                                                                                                                               |
| ------------------------- | -------- | ------- | ----------------------------------------------------------------------------------------------------------------------------------- |
| `COMPANIES_HOUSE_API_KEY` | yes      | —       | Get one at [developer.company-information.service.gov.uk](https://developer.company-information.service.gov.uk/manage-applications) |
| `PORT`                    | no       | `8080`  |                                                                                                                                     |
| `REDIS_URL`               | no       | —       | Redis connection URL; omit to use in-memory cache                                                                                   |
| `CACHE_TTL`               | no       | `24h`   | TTL for cached graph responses                                                                                                      |
| `LOG_LEVEL`               | no       | `INFO`  | `DEBUG`, `INFO`, `WARN`, `ERROR`                                                                                                    |
| `RATE_LIMIT_RPS`          | no       | `2`     | Sustained Companies House API calls per second                                                                                      |
| `RATE_LIMIT_BURST`        | no       | `10`    | Token-bucket burst                                                                                                                  |
| `GEMI_API_KEY`            | no       | —       | Set to any non-empty value to enable the experimental `gr` (GEMI) provider                                                          |

## Provider architecture

Each supported jurisdiction lives under `internal/providers/{cc}/` and implements a common interface:

```go
type Provider interface {
    SearchCompanies(ctx, query) ([]SearchResult, error)
    CompanyGraph(ctx, companyID, depth) (*graph.Response, error)
    OfficerGraph(ctx, officerID) (*graph.Response, error)
}
```

Currently supported:

| Code | Registry                                                                                                                | Notes                                                                                   |
| ---- | ----------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------- |
| `uk` | [Companies House](https://developer-specs.company-information.service.gov.uk/companies-house-public-data-api/reference) | Rate-limited, retrying HTTP client                                                      |
| `gr` | [GEMI Open Data](https://www.businessregistry.gr/publicity/index.xhtml)                                                 | **Experimental** — Greek General Commercial Registry; enabled by setting `GEMI_API_KEY` |

## Adding a new jurisdiction

1. Create `internal/providers/{cc}/` with `client.go`, `types.go`, and `provider.go`.
2. Register it in `cmd/api/main.go`: `reg.Register("xx", xx.New(xx.NewClient(...)))`.
3. Add any required API key env vars to `internal/config/config.go`.

## Testing

```bash
go test ./...
```

CI runs tests with a live Redis container. See `.github/workflows/` at the repo root.
