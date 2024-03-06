# Fusion

Fusion is an RSS aggregator and reader with:

- Lightweight, high performance, easy to deploy
- Support RSS, Atom, JSON feeds
- Import/Export OPML
- Feed groups

## Installation

### 1. Docker

```shell
docker run -it -d -p 8080:8080 -v $(pwd)/fusion:/data \
      -e PASSWORD="123456" \
      rook1e404/fusion
```

Or you can build docker image from scratch:

```shell
docker build -t rook1e404/fusion .
```

### 2. Pre-build

Download an release, then run:

```shell
./fusion-server
```

### 3. Build from source

1. Prepare dependencies

```shell
go mod tidy
cd frontend && npm i
```

2. Build

```shell
./scripts.sh build
```

3. Deploy

```shell
cd build

# edit .env

# run
./fusion-server
```

## ToDo

- Bookmark
- PWA

## Credits

- Frontend is built with: [Sveltekit](https://github.com/sveltejs/kit), [shadcn-svelte](https://github.com/huntabyte/shadcn-svelte)
- Backend is built with: [Echo framework](https://github.com/labstack/echo), [GORM](https://github.com/go-gorm/gorm)
- Parsing feed with [gofeed](https://github.com/mmcdole/gofeed)
