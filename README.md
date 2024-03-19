# Fusion

![preview](./assets/screenshot.png)

Fusion is an RSS aggregator and reader with:

- Lightweight, high performance, easy to deploy
- Support RSS, Atom, JSON feeds
- Group, Bookmark
- Import/Export OPML
- Light/Dark mode
- PWA

## Installation

### 1. Docker

```shell
docker run -it -d -p 8080:8080 -v $(pwd)/fusion:/data \
      -e PASSWORD="123456" \
      rook1e404/fusion
```

<details>
  <summary><b>Other methods</b></summary>

### 2. Pre-build binary

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

</details>

## Credits

- Frontend is built with: [Sveltekit](https://github.com/sveltejs/kit), [shadcn-svelte](https://github.com/huntabyte/shadcn-svelte)
- Backend is built with: [Echo](https://github.com/labstack/echo), [GORM](https://github.com/go-gorm/gorm)
- Parsing feed with [gofeed](https://github.com/mmcdole/gofeed)
