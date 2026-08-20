# Diktyon

**Diktyon** is an open-source tool for investigating UK companies — search any company and explore its officers and persons with significant control (PSCs) as an interactive graph. Polaroid-card nodes, red-thread edges, corkboard aesthetic. [MIT licensed](LICENSE).

- [Diktyon](#diktyon)
  - [Data \& Privacy](#data--privacy)
  - [Features](#features)
  - [Stack](#stack)
  - [Getting Started](#getting-started)
  - [Development](#development)
  - [API](#api)
  - [Examples](#examples)
  - [Configuration](#configuration)


## Data & Privacy

All company data is sourced from the [Companies House public API](https://developer-specs.company-information.service.gov.uk/companies-house-public-data-api/reference). No cookies, no tracking, no data collected. ☕ [Buy me a coffee](https://ko-fi.com/angelospanag) if you find it useful.

## Features

- Interactive graph of UK company networks (officers, PSCs)
- Click any node to expand into its connections
- Depth-2 expansion — follow an officer to their other current companies
- 24-hour response cache with Redis (in-memory fallback when Redis is unavailable)
- Rate-limited Companies House client (token bucket, 600 req / 5 min)
- OpenAPI spec auto-generated from Go handler types via [Huma](https://huma.rocks); TypeScript client generated from the spec via [hey-api](https://heyapi.dev)
- Experimental support for GEMI (Greece General Commercial Registry) — opt-in via `GEMI_API_KEY`

## Stack

| Directory     | Purpose                               | Technologies                                                                                 |
| ------------- | ------------------------------------- | -------------------------------------------------------------------------------------------- |
| `api/`        | Companies House proxy + graph builder | Go 1.27, Huma v2, chi, go-redis                                                              |
| `ui/`         | Interactive corkboard graph           | Next.js 16, React 19, TypeScript 5, Tailwind CSS v4, d3-force v3, TanStack Query v5, hey-api |
| `compose.yml` | Full-stack orchestration              | Podman Compose, Redis 8                                                                      |

## Getting Started

[mise](https://mise.jdx.dev/) manages the pinned toolchain.
[Podman](https://podman.io/) is a separate prerequisite (install Podman Desktop or equivalent; version 4+ for the built-in `podman compose`).

```bash
# macOS / Linux
curl https://mise.run | sh

# Windows
winget install jdx.mise
```

Activate mise in your shell (`~/.zshrc`):

```zsh
eval "$(mise activate zsh)"
```

Then, in the repo:

```bash
mise trust    # one-time, confirms you trust this repo's mise.toml
mise install  # downloads Go, Bun, and golangci-lint
```

Get a free Companies House API key at
[developer.company-information.service.gov.uk](https://developer.company-information.service.gov.uk/manage-applications),
then:

```bash
cp api/.env.example api/.env
# edit api/.env and set COMPANIES_HOUSE_API_KEY
```

API: **http://localhost:8080** • UI: **http://localhost:3000** • Docs: **http://localhost:8080/docs**

## Development

Run the API and UI in separate terminals:

```bash
# Terminal 1
mise -C api run dev

# Terminal 2
mise -C ui run install   # first time only
mise -C ui run dev
```

The Next.js dev server proxies `/api/*` to `:8080` — no CORS config needed.

Root `mise.toml` exposes cross-project aggregates and Podman tasks. Service-specific tasks run from inside `api/` or `ui/` (or via `mise -C <dir>`).

| Command                 | Description                                      |
| ----------------------- | ------------------------------------------------ |
| `mise run fmt`          | Format all code (Go + TypeScript)                |
| `mise run lint`         | Lint all code (Go + TypeScript)                  |
| `mise run deps`         | Update all dependencies (Go + Bun)               |
| `mise run generate`     | Regenerate OpenAPI schema then TypeScript client |
| `mise run compose:up`   | Build + start full stack with Podman Compose     |
| `mise run compose:down` | Stop all services                                |
| `mise run compose:logs` | Follow compose logs                              |

| API command (`cd api/`) | Description                           |
| ----------------------- | ------------------------------------- |
| `mise run dev`          | Start the Go API on :8080             |
| `mise run fmt`          | Format Go code                        |
| `mise run lint`         | Lint Go code                          |
| `mise run test`         | Run Go tests with race detector       |
| `mise run vuln`         | Scan dependencies for vulnerabilities |
| `mise run tidy`         | Tidy Go module dependencies           |
| `mise run build`        | Compile a static binary               |

| UI command (`cd ui/`) | Description                           |
| --------------------- | ------------------------------------- |
| `mise run dev`        | Start the Next.js dev server on :3000 |
| `mise run fmt`        | Format TypeScript code                |
| `mise run lint`       | Lint TypeScript code                  |
| `mise run typecheck`  | TypeScript type check                 |
| `mise run build`      | Production Next.js build              |

## API

OpenAPI docs available at `http://localhost:8080/docs`.

All endpoints accept an optional `?country=` parameter (default: `uk`).

| Method | Endpoint                              | Description                                            |
| ------ | ------------------------------------- | ------------------------------------------------------ |
| GET    | `/api/search?q=&country=`             | Search companies by name (across supported registries) |
| GET    | `/api/company/{number}/graph?depth=1` | Company graph (depth 1 or 2)                           |
| GET    | `/api/officer/{id}/graph`             | All current appointments for an officer                |

## Examples

```bash
# Search for a company
curl "http://localhost:8080/api/search?q=apple"

# Company graph — direct connections (officers + PSCs)
curl "http://localhost:8080/api/company/00000001/graph"

# Company graph — depth 2 (also expands each officer's other companies)
curl "http://localhost:8080/api/company/00000001/graph?depth=2"

# All current appointments for an officer
curl "http://localhost:8080/api/officer/{officer_id}/graph"
```

## Configuration

```bash
# api/.env
COMPANIES_HOUSE_API_KEY=""   # required — get at developer.company-information.service.gov.uk
PORT="8080"
REDIS_URL=""                 # optional — omit to use in-memory cache
CACHE_TTL="24h"
LOG_LEVEL="INFO"
RATE_LIMIT_RPS="2"
RATE_LIMIT_BURST="10"
GEMI_API_KEY=""              # optional — set to any non-empty value to enable experimental GEMI (Greece) support
```
