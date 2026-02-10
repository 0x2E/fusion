#!/bin/sh

set -eu

resolve_version() {
  if [ -n "${FUSION_VERSION:-}" ]; then
    printf '%s\n' "$FUSION_VERSION"
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
  echo "Using fusion version string: ${version}"

  (
    cd frontend
    pnpm install --frozen-lockfile --prefer-offline
    VITE_FUSION_VERSION="$version" pnpm run build
  )
}

build_backend() {
  target_os=${1:-$(go env GOOS)}
  target_arch=${2:-$(go env GOARCH)}
  root=$(pwd)
  output_path=${3:-"${root}/build/fusion"}
  echo "building backend for OS: ${target_os}, Arch: ${target_arch}, Output: ${output_path}"

  mkdir -p "$(dirname "$output_path")"
  (
    cd backend
    CGO_ENABLED=0 GOOS=${target_os} GOARCH=${target_arch} go build \
      -trimpath \
      -ldflags '-extldflags "-static"' \
      -o "$output_path" \
      ./cmd/fusion
  )
}

release() {
  echo "building release artifacts"
  rm -rf ./dist
  mkdir -p ./dist

  platforms="linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64"

  for platform in $platforms; do
    os=${platform%/*}
    arch=${platform#*/}
    echo "--- building ${os}/${arch} ---"

    bin_name="fusion"
    if [ "$os" = "windows" ]; then
      bin_name="fusion.exe"
    fi

    build_backend "$os" "$arch" "./dist/${bin_name}"

    if [ "$os" = "linux" ]; then
      cp "./dist/${bin_name}" "./dist/fusion-linux-${arch}"
    fi

    archive="fusion_${os}_${arch}.zip"
    zip -j "./dist/${archive}" "./dist/${bin_name}" LICENSE* README* || \
      zip -j "./dist/${archive}" "./dist/${bin_name}"
    rm "./dist/${bin_name}"
  done

  (
    cd ./dist
    sha256sum ./*.zip ./fusion-linux-* > checksums.txt
  )

  echo "release artifacts:"
  ls -lh ./dist/
}

build() {
  test_backend
  build_frontend
  build_backend
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
*)
  usage
  exit 1
  ;;
esac
