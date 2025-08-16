<h1 align="center">Fusion</h1>
<p align="center">A lightweight RSS feed aggregator and reader.</p>
<p align="center">
  <a href="https://news.ycombinator.com/item?id=40522244">
    <img
      alt="Featured on Hacker News"
      src="https://hackerbadge.vercel.app/api?id=40522244"
    />
  </a>
</p>

![screenshots](./assets/screenshots.png)

## Features

- Group, bookmark, search, automatic feed sniffing, OPML file import/export
- Supports RSS, Atom, and JSON feed types
- Responsive, dark mode, PWA, keyboard shortcuts
- Lightweight and self-hosted friendly
  - Built with Golang and SQLite; deploys with a single binary or a Docker container
  - Uses about 80MB of memory
- Internationalization (i18n): English, Chinese, German, French, Spanish, Russian, Portuguese, and Swedish

## Installation

<details>
<summary>Docker</summary>

> Use `latest` tag for the latest release version.
>
> Use `main` tag for the latest development version.

- Docker CLI

```shell
docker run -it -d -p 8080:8080 \
  -v $(pwd)/fusion:/data \
  -e PASSWORD="fusion" \
  ghcr.io/0x2e/fusion:latest
```

- Docker Compose

```yaml
version: "3"
services:
  fusion:
    image: ghcr.io/0x2e/fusion:latest
    ports:
      - "127.0.0.1:8080:8080"
    environment:
      - PASSWORD=fusion
    restart: "unless-stopped"
    volumes:
      # Change `./data` to where you want the files stored
      - ./data:/data
```

</details>

<details>
<summary>Pre-built binary</summary>

Download from [Releases](https://github.com/0x2E/fusion/releases).
</details>

<details>
  <summary>One-Click Deployment</summary>

[Deploy on Fly.io](./fly.toml)

[![Deploy on Zeabur](https://zeabur.com/button.svg)](https://zeabur.com/templates/7FRK0K?referralCode=rook1e404)

Maintained by community:

[![Deploy on Railway](https://railway.com/button.svg)](https://railway.com/template/XSPFK0?referralCode=milo)

</details>

<details>
  <summary>Build from source</summary>

  Check out the "Contributing" section.
</details>

## Configuration

All configuration items can be found in [`.env.example`](./.env.example).

Fusion can be configured in many ways:

- System environment variables, such as those set by `export PASSWORD=123abc`.
- Create a `.env` file in the same directory as the binary. Note that values in `.env` file can be overwritten by system environment variables.

## Contributing

Contributions are welcome! Before contributing, please read the [Contributing Guidelines](./CONTRIBUTING.md).

- Prepare environment: Go 1.24+, Node.js 24+ (and pnpm).
- Check out the commands in `scripts.sh`.

For example:

```shell
./scripts.sh build
```

## Credits

- Front-end is built with: [Sveltekit](https://github.com/sveltejs/kit), [daisyUI](https://github.com/saadeghi/daisyui)
- Back-end is built with: [Echo](https://github.com/labstack/echo), [GORM](https://github.com/go-gorm/gorm)
- Parsing feed with [gofeed](https://github.com/mmcdole/gofeed)
- Logo by [Icons8](https://icons8.com/icon/FeQbTvGTsiN5/news)
