# Fusion

![preview](./assets/screenshot.png)

Fusion is an RSS aggregator and reader with:

- Group, Bookmark, Search, Sniff feeds automatically, Import/Export OPML file
- Support RSS, Atom, JSON types feed
- Responsive, Light/Dark mode, PWA
- Lightweight, Self-hosted friendly
  - Built with Golang and SQLite, Deploy with a single binary
  - Pre-built Docker image
  - Uses about 80MB of memory

## Installation

### 1. Docker

```shell
docker run -it -d -p 8080:8080 -v $(pwd)/fusion:/data \
      -e PASSWORD="123456" \
      rook1e404/fusion
```

### 2. Pre-built binary

Download from [Releases](https://github.com/0x2E/fusion/releases).

### 3. Build from source

1. Prepare dependencies: Go 1.22+, Node.js 21+.
2. Check `scripts.sh` for more details.

For example:

```shell
./scripts.sh build
```

## Configuration

Fusion can be configured in many ways:

- System environment variables, such as those set by `export PASSWORD=123abc`.
- Create a `.env` file in the same directory as the binary file, and then copy the items you want to modify into it.
  - NOTE: values in `.env` file can be overwritten by system environment variables.

All configuration items can be found [here](https://github.com/0x2E/fusion/blob/main/.env).

## Credits

- Front-end is built with: [Sveltekit](https://github.com/sveltejs/kit), [shadcn-svelte](https://github.com/huntabyte/shadcn-svelte)
- Back-end is built with: [Echo](https://github.com/labstack/echo), [GORM](https://github.com/go-gorm/gorm)
- Parsing feed with [gofeed](https://github.com/mmcdole/gofeed)
