<h1 align="center">Fusion</h1>
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
  -v $(pwd)/fusion:/data \
  -e FUSION_PASSWORD="fusion" \
  ghcr.io/0x2e/fusion:latest
```

Open `http://localhost:8080`.

Docker Compose example:

```yaml
version: "3"
services:
  fusion:
    image: ghcr.io/0x2e/fusion:latest
    ports:
      - "127.0.0.1:8080:8080"
    environment:
      - FUSION_PASSWORD=fusion
    restart: unless-stopped
    volumes:
      - ./data:/data
```

## Other Installation Options

- Pre-built binary: download from [Releases](https://github.com/0x2E/fusion/releases)
- Build from source: see [Contributing](./CONTRIBUTING.md)
- One-click deployment:
  - [Deploy on Fly.io](./fly.toml)
  - [Deploy on Railway](https://railway.com/template/XSPFK0?referralCode=milo) (community maintained)

## Configuration

All config keys are documented in [`.env.example`](./.env.example).

Common keys:

- `FUSION_PASSWORD` (required unless `FUSION_ALLOW_EMPTY_PASSWORD=true`)
- `FUSION_PORT` (default `8080`)
- `FUSION_PULL_INTERVAL`, `FUSION_PULL_TIMEOUT`, `FUSION_PULL_CONCURRENCY`, `FUSION_PULL_MAX_BACKOFF`
- `FUSION_CORS_ALLOWED_ORIGINS`, `FUSION_TRUSTED_PROXIES`
- `FUSION_OIDC_*` for optional SSO

Legacy env names (`DB`, `PASSWORD`, `PORT`) are still accepted for backward compatibility.

## Database

Fusion supports SQLite (default) and PostgreSQL. The two are mutually exclusive — set one or the other.

### SQLite (default)

No extra setup required. Set the file path with `FUSION_DB_PATH` (default: `fusion.db`).

```shell
docker run -it -d -p 8080:8080 \
  -v $(pwd)/fusion:/data \
  -e FUSION_PASSWORD="fusion" \
  -e FUSION_DB_PATH="/data/fusion.db" \
  ghcr.io/0x2e/fusion:latest
```

### PostgreSQL

Set `FUSION_DATABASE_URL` to a PostgreSQL connection string. When this variable is present, `FUSION_DB_PATH` is ignored and Fusion connects to PostgreSQL instead.

```shell
docker run -it -d -p 8080:8080 \
  -e FUSION_PASSWORD="fusion" \
  -e FUSION_DATABASE_URL="postgres://user:password@host:5432/fusion?sslmode=disable" \
  ghcr.io/0x2e/fusion:latest
```

Docker Compose example with a managed PostgreSQL service:

```yaml
services:
  postgres:
    image: postgres:17
    environment:
      POSTGRES_DB: fusion
      POSTGRES_USER: fusion
      POSTGRES_PASSWORD: fusion
    volumes:
      - postgres_data:/var/lib/postgresql/data

  fusion:
    image: ghcr.io/0x2e/fusion:latest
    environment:
      FUSION_PASSWORD: changeme
      FUSION_DATABASE_URL: postgres://fusion:fusion@postgres:5432/fusion?sslmode=disable
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
