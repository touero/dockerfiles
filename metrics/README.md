# Docker Hub pull metrics

Daily Docker Hub pull-count tracker for the images built from this repo.
Written in Go, single static binary, no runtime dependencies.

## Layout

```
metrics/
├── cmd/collect        # fetch pull counts -> CSV
├── cmd/plot           # CSV -> SVG trend charts
├── internal/chart     # dependency-free modern SVG renderer
├── internal/hub       # Docker Hub v2 API client
├── internal/store     # CSV load/save/upsert
├── data/pull_counts.csv
└── go.mod
```

No third-party dependencies: everything uses the Go standard library.

Data source: the public Docker Hub v2 API
(`GET /v2/repositories/{namespace}/{repo}/`), which needs no token for public
repositories.

## How it works

1. `collect` reads `pull_count` / `star_count` for each repo and appends today's
   snapshot to `data/pull_counts.csv` (idempotent — re-running on the same day
   replaces the row).
2. `plot` renders modern SVG charts into `docs/`, each in a light and a dark
   variant:
   - `pull-counts.svg` / `pull-counts-dark.svg` — cumulative pulls (log scale)
   - `pull-counts-daily.svg` / `pull-counts-daily-dark.svg` — new pulls per day
3. [`.github/workflows/metrics.yml`](../.github/workflows/metrics.yml) runs both
   every day at 02:17 UTC and commits the results back to the repository.

The CSV uses a "long" format so adding or removing a repo never changes the
schema:

```
date,repo,pull_count,star_count
2025-01-01,fileserver,100,2
2025-01-01,gitweb,50,1
```

## Local usage

```bash
cd metrics
go run ./cmd/collect   # needs DOCKERHUB_USERNAME
go run ./cmd/plot
```

Environment variables:

| Variable | Default | Used by |
|---|---|---|
| `DOCKERHUB_USERNAME` / `DOCKERHUB_NAMESPACE` | — | collect |
| `REPOS` | `fileserver gitweb wechat` | collect |
| `DATA_FILE` | `metrics/data/pull_counts.csv` | both |
| `OUT_DIR` | `docs` | plot |

Paths are relative to the repository root, so run the built binaries from there
(or override the variables when running from `metrics/`).
