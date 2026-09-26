# mm-budget

Household budget app for two people and a joint account. Log expenses in seconds from your phone, then see where the money goes, month by month.

- **Quick add**: type the amount, tap a category, save. The Add screen is the landing page, and you can install the app to your home screen as a PWA.
- **Month view**: income, expenses, investments, leftover and savings rate. Also budgets with progress bars, insights, a category breakdown you can drill into, and an inbox of pending recurring entries to confirm.
- **Charts**: 12-month trends by category, income vs expenses vs savings, and year-over-year comparison.
- **Entries**: search and filter by text, type, category, payer, tag, amount and date, with totals.
- **Recurring templates**: monthly entries (mortgage, subscriptions, salary, ETF contributions) created automatically on the 1st. Variable ones (electricity, salary) arrive as *pending* estimates for you to confirm.
- **Settings**: people, categories (icons and colors, archive), templates, budgets, CSV export (`;` separator, decimal comma) and backup download.
- **Backups**: a nightly `VACUUM INTO` snapshot goes to `data/backups/`, and the last 30 are kept.

There is no authentication. Run it on a trusted network (LAN or VPN).

## Run with Docker

```sh
cp .env.example .env        # adjust PORT, DATA_LOCATION, TZ
mkdir -p data && sudo chown 65532:65532 data   # container runs as non-root 65532 (or set PUID/PGID)
make up                     # docker compose up -d --build
```

Open `http://<host>:8080`.

Behind a reverse proxy that uses an external Docker network called `proxy`, no port is published:

```sh
make up-proxy               # compose.yaml + compose.proxy.yaml
```

Then point the proxy at `mm-budget:8080`.

| Env | Default | |
|---|---|---|
| `PORT` | `8080` | Host port (base compose only) |
| `DATA_LOCATION` | `./data` | Host dir for `mm-budget.db` and `backups/` |
| `PUID` / `PGID` | `65532` | User the container runs as. Must own `DATA_LOCATION` |
| `TZ` | `Europe/Lisbon` | Defines "today" and recurring generation |
| `LOG_LEVEL` | `info` | `debug` logs every request |

To restore a backup, stop the app and replace `data/mm-budget.db` with a snapshot from `data/backups/`. Then start it again.

## Develop

Requirements: Go 1.27, Node 24, [sqlc](https://sqlc.dev), [gofumpt](https://github.com/mvdan/gofumpt).

```sh
make dev-api     # Go API on :8080 (data in ./data)
make dev-web     # Vite on :5173, proxies /api to :8080
make test        # go test -race ./... + vitest
make generate    # sqlc after editing internal/store/queries/*.sql
make e2e         # Playwright smoke test (expects the app on :8080)
```

Layout:

```
cmd/mm-budget        main: HTTP server, recurring + backup jobs, healthcheck subcommand
internal/api         JSON API (docs/API.md) + SPA serving
internal/store       SQLite, goose migrations (embedded), sqlc queries
internal/recurring   monthly template generation
internal/reports     month/trends/year aggregation + insights
internal/export      CSV export
internal/backup      VACUUM INTO snapshots
web/                 React + Vite + Tailwind + shadcn/ui SPA (embedded into the binary)
docs/                SPEC.md (product spec), API.md (HTTP contract)
```
