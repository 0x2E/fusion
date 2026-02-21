#!/bin/sh

set -eu

resolve_version() {
  if [ -n "${REEDME_VERSION:-}" ]; then
    printf '%s\n' "$REEDME_VERSION"
    return
  fi

  if git describe --tags --abbrev=0 >/dev/null 2>&1; then
    git describe --tags --abbrev=0
    return
  fi

  git rev-parse --short HEAD
}

test_backend() {
  echo "testing backend"
  (cd backend && go test ./...)
}

build_frontend() {
  echo "building frontend"
  version=$(resolve_version)
  echo "Using reedme version string: ${version}"

  (
    cd frontend
    pnpm install --frozen-lockfile --prefer-offline
    VITE_REEDME_VERSION="$version" pnpm run build
  )

  echo "syncing frontend build artifacts for backend embed"
  rm -rf backend/internal/web/dist
  mkdir -p backend/internal/web/dist
  cp -R frontend/dist/. backend/internal/web/dist/
  printf '%s\n' "This file keeps the embedded dist directory in version control." > backend/internal/web/dist/.keep
}

build_backend() {
  target_os=${1:-$(go env GOOS)}
  target_arch=${2:-$(go env GOARCH)}
  root=$(pwd)
  output_path=${3:-"${root}/build/reedme"}

  case "$output_path" in
  /*) ;;
  *) output_path="${root}/${output_path#./}" ;;
  esac

  if [ ! -f backend/internal/web/dist/index.html ]; then
    echo "frontend build artifacts not found for embed"
    echo "run ./scripts.sh build-frontend before building backend"
    exit 1
  fi

  echo "building backend for OS: ${target_os}, Arch: ${target_arch}, Output: ${output_path}"

  mkdir -p "$(dirname "$output_path")"
  (
    cd backend
    CGO_ENABLED=0 GOOS=${target_os} GOARCH=${target_arch} go build \
      -trimpath \
      -ldflags '-extldflags "-static"' \
      -o "$output_path" \
      ./cmd/reedme
  )
}

release() {
  echo "building release artifacts"
  rm -rf ./dist
  mkdir -p ./dist

  build_frontend

  platforms="linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64"

  for platform in $platforms; do
    os=${platform%/*}
    arch=${platform#*/}
    echo "--- building ${os}/${arch} ---"

    bin_name="reedme"
    if [ "$os" = "windows" ]; then
      bin_name="reedme.exe"
    fi

    build_backend "$os" "$arch" "./dist/${bin_name}"

    if [ "$os" = "linux" ]; then
      cp "./dist/${bin_name}" "./dist/reedme-linux-${arch}"
    fi

    archive="reedme_${os}_${arch}.zip"
    zip -j "./dist/${archive}" "./dist/${bin_name}" LICENSE* README* || \
      zip -j "./dist/${archive}" "./dist/${bin_name}"
    rm "./dist/${bin_name}"
  done

  (
    cd ./dist
    sha256sum ./*.zip ./reedme-linux-* > checksums.txt
  )

  echo "release artifacts:"
  ls -lh ./dist/
}

build() {
  test_backend
  build_frontend
  build_backend
}

setup_hooks() {
  git config core.hooksPath .githooks
  echo "git hooks configured (using .githooks/)"
}

usage() {
  cat <<'EOF'
Usage: ./scripts.sh <command>

Commands:
  test-backend             Run backend tests
  build-frontend           Build frontend bundle
  build-backend [os] [arch] [output]
                           Build backend binary
  build                    Run backend tests and build all
  release                  Build release archives and checksums
  setup-hooks              Install git hooks (commit-msg linting)
EOF
}

case "${1:-}" in
"test" | "test-backend")
  test_backend
  ;;
"build-frontend")
  build_frontend
  ;;
"build-backend")
  build_backend "${2:-}" "${3:-}" "${4:-}"
  ;;
"build")
  build
  ;;
"release")
  release
  ;;
"setup-hooks")
  setup_hooks
  ;;
*)
  usage
  exit 1
  ;;
esac
