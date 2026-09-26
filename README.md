# 🐷 Mealheiro

> Household budget for two people and a joint account: log expenses in seconds, see where the money goes.

[![release](https://github.com/melojms/mealheiro/actions/workflows/release.yml/badge.svg)](https://github.com/melojms/mealheiro/actions/workflows/release.yml)
[![GitHub release](https://img.shields.io/github/v/release/melojms/mealheiro)](https://github.com/melojms/mealheiro/releases/latest)
[![GHCR image](https://img.shields.io/badge/ghcr.io-melojms%2Fmealheiro-2496ED?logo=docker&logoColor=white)](https://github.com/melojms/mealheiro/pkgs/container/mealheiro)
[![Go version](https://img.shields.io/github/go-mod/go-version/melojms/mealheiro)](go.mod)

- ⚡ **Quick add**: type the amount, tap a category, save. The Add screen is the landing page.
- 📅 **Month view**: income, expenses, investments, leftover and savings rate, a category breakdown you can drill into, and an inbox of pending recurring entries to confirm.
- 📈 **Charts**: 12-month trends by category, income vs expenses vs savings, and year-over-year comparison.
- 🔎 **Entries**: search and filter by text, type, category, payer, tag, amount and date, with totals.
- 🔁 **Recurring templates**: monthly entries (mortgage, subscriptions, salary, ETF contributions) created automatically on the 1st. Variable ones (electricity, salary) arrive as *pending* estimates for you to confirm.
- 🎯 **Budgets**: monthly limits per category and overall, with progress bars.
- 💡 **Insights**: highlights for the month in the month view.
- 📤 **CSV export**: `;` separator and decimal comma, filtered by date range and type.
- 💾 **Backups**: a nightly `VACUUM INTO` snapshot goes to `data/backups/` (last 30 kept), plus a download button in Settings.
- 🌙 **Dark mode**: light, dark, or follow the system.
- 📱 **PWA**: install it to your phone's home screen.

Settings also covers people, categories (icons and colors, archive) and templates.

## ⚠️ Security note

There is no authentication. Run it on a trusted network (LAN or VPN).

## 🚀 Run with Docker

Just mount a folder. It can be missing, empty, root-owned or left over from an older version: the container fixes ownership itself, then drops to `PUID:PGID` (default `1000:1000`). No `chown` needed.

Minimal `compose.yaml`:

```yaml
services:
  mealheiro:
    image: ghcr.io/melojms/mealheiro:latest
    restart: unless-stopped
    environment:
      TZ: Europe/Lisbon
    ports:
      - "7447:7447"
    volumes:
      - ./data:/data
```

Open `http://<host>:7447`.

From a clone of this repo, the bundled [`compose.yaml`](compose.yaml) adds optional hardening: read-only root filesystem, `no-new-privileges`, and all capabilities dropped except the four needed for the startup chown and privilege drop (explained inline):

```sh
cp .env.example .env        # optional: HOST_PORT, DATA_LOCATION, PUID/PGID, TZ
make up                     # build from source; or `make pull-up` for the released image
```

Behind a reverse proxy that uses an external Docker network called `proxy`, no port is published:

```sh
make up-proxy               # compose.yaml + compose.proxy.yaml
```

Then point the proxy at `mealheiro:$PORT` (`mealheiro:7447` by default).

| Env | Default | |
|---|---|---|
| `HOST_PORT` | `7447` | Host port the app is published on (base compose only) |
| `PORT` | `7447` | Port the server listens on inside the container (1-65535) |
| `DATA_LOCATION` | `./data` | Host dir for `mealheiro.db` and `backups/` |
| `PUID` / `PGID` | `1000` | Owner of the files in `DATA_LOCATION`; the container fixes permissions itself and runs as this user |
| `TZ` | `Europe/Lisbon` | Defines "today" and recurring generation |
| `LOG_LEVEL` | `info` | `debug` logs every request |

To restore a backup, stop the app and replace `data/mealheiro.db` with a snapshot from `data/backups/`. Then start it again.

## 🏷️ Releases & deploy

Every merge to `main` runs [`.github/workflows/release.yml`](.github/workflows/release.yml): tests, then a `vX.Y.Z` tag, a GitHub Release with generated notes, and an image on GHCR.

- **Version**: computed from conventional commits since the last tag. `feat` bumps minor, breaking (`!` or `BREAKING CHANGE:`) bumps major (minor while on 0.x), anything else bumps patch. The first release is `v0.1.0`.
- **Image tags**: `ghcr.io/melojms/mealheiro:X.Y.Z`, `:X.Y`, `:latest`, `:sha-<short>`.

On the homelab server:

```sh
# .env: VERSION=0.1.0 to pin a release, or VERSION=latest
make pull-up                # docker compose pull && up -d (no local build)
```

The GHCR package is private by default. Either make it public (package settings, "Change visibility"), or log in once on the server with a PAT that has `read:packages`:

```sh
echo "$PAT" | docker login ghcr.io -u melojms --password-stdin
```

`make up` still builds the image locally from source.

## 🛠️ Develop

Requirements: Go 1.27, Node 24, [sqlc](https://sqlc.dev), [gofumpt](https://github.com/mvdan/gofumpt).

```sh
make dev-api     # Go API on :7447 (data in ./data)
make dev-web     # Vite on :5173, proxies /api to :7447
make test        # go test -race ./... + vitest
make generate    # sqlc after editing internal/store/queries/*.sql
make e2e         # Playwright smoke test (expects the app on :7447)
```

## 🗂️ Layout

```
cmd/mealheiro        main: HTTP server, recurring + backup jobs, healthcheck subcommand
internal/api         JSON API (docs/API.md) + SPA serving
internal/store       SQLite, goose migrations (embedded), sqlc queries
internal/recurring   monthly template generation
internal/reports     month/trends/year aggregation + insights
internal/export      CSV export
internal/backup      VACUUM INTO snapshots
internal/privdrop    startup chown of the data dir + drop to PUID:PGID
web/                 React + Vite + Tailwind + shadcn/ui SPA (embedded into the binary)
docs/                SPEC.md (product spec), API.md (HTTP contract)
```
