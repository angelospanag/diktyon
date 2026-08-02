# Diktyon

UK corporate intelligence tool. Search any UK company and explore its network of officers and persons with significant control (PSCs) as an interactive corkboard graph. Experimental support for GEMI (Greece General Commercial Registry) via a pluggable provider architecture (`api/internal/providers/`) — opt-in via `GEMI_API_KEY`.

Target audience: journalists, property investigators, researchers — not compliance teams. Aesthetic: detective's conspiracy corkboard (Polaroid cards, red-thread edges, cork texture).

## Repo layout

```
diktyon/
├── api/              Go backend (Companies House API proxy + graph builder)
│   └── openapi.yaml  Generated OpenAPI spec (`mise -C api run schema`)
├── ui/               Next.js 16 frontend (d3-force layout, SVG edges, HTML cards)
│   └── client/       Generated TypeScript client (`mise run generate`)
├── compose.yml       Full-stack Podman Compose (backend + ui + Redis)
└── .github/          CI (lint · test · typecheck · build)
```

See `api/CLAUDE.md` and `ui/CLAUDE.md` for per-service details.

## Running the full stack

```bash
# Requires: Podman, api/.env with COMPANIES_HOUSE_API_KEY set (see api/.env.example)
podman compose up
# API → http://localhost:8080
# Frontend → http://localhost:3000
```

For local development, run each service separately — see their CLAUDE.md files.

## CI (GitHub Actions)

Five jobs on push/PR to `main`:
- **lint** — golangci-lint on the Go code
- **lint-frontend** — Biome check on the TypeScript code
- **test** — `go test ./...` with a live Redis service container
- **typecheck-frontend** — `tsc --noEmit`
- **build** — Podman image builds (depends on the four above)

Task runner is `mise` — see `mise.toml` for task definitions.

## Key constraints

- `api/.env` is gitignored. Never commit it or any file containing `COMPANIES_HOUSE_API_KEY`.
- The Companies House public API is rate-limited to 600 req/5 min per key. The Go client enforces this with a token-bucket limiter — don't bypass it.
- Node IDs are stable across sessions (`company:{number}`, `officer:{id}`, `psc:{sanitized_name}`). The frontend derives each Polaroid card's tilt from a hash of the node ID (`ui/lib/graph.ts`'s `tiltAngle`).
