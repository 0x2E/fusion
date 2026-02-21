<h1 align="center">ReedMe</h1>
<p align="center">A lightweight RSS reader.</p>

<p align="center">
  <img src="./assets/article_list_light.png" alt="Article list view" width="48.5%" />&nbsp;
  <img src="./assets/article_detail_light.png" alt="Article detail view" width="48.5%" />
</p>

## Features

- Fast reading workflow: unread tracking, bookmarks, search, and Google Reader-style keyboard shortcuts
- Feed management: RSS/Atom parsing, feed auto-discovery, and group organization
- Responsive web UI with PWA support
- Self-hosting friendly: single binary or Docker deployment
- Built-in i18n: English, Chinese, German, French, Spanish, Russian, Portuguese, Swedish
- No AI features by design: focused, distraction-free RSS reading

## Quick Start (Docker)

> `latest` is the latest release.
>
> `main` is the latest development build.

```shell
docker run -it -d -p 8080:8080 \
  -v $(pwd)/reedme:/data \
  -e REEDME_PASSWORD="reedme" \
  ghcr.io/patrickjmcd/reedme:latest
```

Open `http://localhost:8080`.

Docker Compose example:

```yaml
version: "3"
services:
  reedme:
    image: ghcr.io/patrickjmcd/reedme:latest
    ports:
      - "127.0.0.1:8080:8080"
    environment:
      - REEDME_PASSWORD=reedme
    restart: unless-stopped
    volumes:
      - ./data:/data
```

## Other Installation Options

- Pre-built binary: download from [Releases](https://github.com/patrickjmcd/reedme/releases)
- Build from source: see [Contributing](./CONTRIBUTING.md)
- One-click deployment:
  - [Deploy on Fly.io](./fly.toml)
  - [Deploy on Railway](https://railway.com/template/XSPFK0?referralCode=milo) (community maintained)

## Configuration

All config keys are documented in [`.env.example`](./.env.example).

Common keys:

- `REEDME_PASSWORD` (required unless `REEDME_ALLOW_EMPTY_PASSWORD=true`)
- `REEDME_PORT` (default `8080`)
- `REEDME_PULL_INTERVAL`, `REEDME_PULL_TIMEOUT`, `REEDME_PULL_CONCURRENCY`, `REEDME_PULL_MAX_BACKOFF`
- `REEDME_CORS_ALLOWED_ORIGINS`, `REEDME_TRUSTED_PROXIES`
- `REEDME_OIDC_*` for optional SSO

Legacy env names (`DB`, `PASSWORD`, `PORT`) are still accepted for backward compatibility.

## Database

ReedMe supports SQLite (default) and PostgreSQL. The two are mutually exclusive — set one or the other.

### SQLite (default)

No extra setup required. Set the file path with `REEDME_DB_PATH` (default: `reedme.db`).

```shell
docker run -it -d -p 8080:8080 \
  -v $(pwd)/reedme:/data \
  -e REEDME_PASSWORD="reedme" \
  -e REEDME_DB_PATH="/data/reedme.db" \
  ghcr.io/patrickjmcd/reedme:latest
```

### PostgreSQL

Set `REEDME_DATABASE_URL` to a PostgreSQL connection string. When this variable is present, `REEDME_DB_PATH` is ignored and ReedMe connects to PostgreSQL instead.

```shell
docker run -it -d -p 8080:8080 \
  -e REEDME_PASSWORD="reedme" \
  -e REEDME_DATABASE_URL="postgres://user:password@host:5432/reedme?sslmode=disable" \
  ghcr.io/patrickjmcd/reedme:latest
```

Docker Compose example with a managed PostgreSQL service:

```yaml
services:
  postgres:
    image: postgres:17
    environment:
      POSTGRES_DB: reedme
      POSTGRES_USER: reedme
      POSTGRES_PASSWORD: reedme
    volumes:
      - postgres_data:/var/lib/postgresql/data

  reedme:
    image: ghcr.io/patrickjmcd/reedme:latest
    environment:
      REEDME_PASSWORD: changeme
      REEDME_DATABASE_URL: postgres://reedme:reedme@postgres:5432/reedme?sslmode=disable
    ports:
      - "127.0.0.1:8080:8080"
    depends_on:
      postgres:
        condition: service_healthy

volumes:
  postgres_data:
```

Migrations run automatically on startup for both databases. There is no migration path between the two backends — pick one and stick with it.

## Documentation

- API contract (OpenAPI): [`docs/openapi.yaml`](./docs/openapi.yaml)
- Backend design: [`docs/backend-design.md`](./docs/backend-design.md)
- Frontend design: [`docs/frontend-design.md`](./docs/frontend-design.md)
- Legacy schema reference (kept for migration work): [`docs/old-database-schema.md`](./docs/old-database-schema.md)

## Development

- Requirements: Go `1.25+`, Node.js `24+`, pnpm
- Helpful commands are in [`scripts.sh`](./scripts.sh)
- Frontend i18n key check: `cd frontend && npm run check:i18n`

Example:

```shell
./scripts.sh build
```

## Contributing

Contributions are welcome. Please read [Contributing Guidelines](./CONTRIBUTING.md) before opening a PR.

## Credits

- Feed parsing powered by [gofeed](https://github.com/mmcdole/gofeed)
