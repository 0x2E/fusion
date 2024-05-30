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

Download an release, edit `.env`, then run:

```shell
./fusion
```

### 3. Build from source

1. Prepare dependencies: Go 1.22, Node 21 with NPM
2. Build

```shell
./scripts.sh build
```

3. Deploy

```shell
cd build

# edit .env

# run
./fusion
```

## Credits

- Front-end is built with: [Sveltekit](https://github.com/sveltejs/kit), [shadcn-svelte](https://github.com/huntabyte/shadcn-svelte)
- Back-end is built with: [Echo](https://github.com/labstack/echo), [GORM](https://github.com/go-gorm/gorm)
- Parsing feed with [gofeed](https://github.com/mmcdole/gofeed)
