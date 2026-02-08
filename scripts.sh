#!/bin/sh

# Exit on first failure.
set -e

test_go() {
  echo "testing"
  # make some files for embed
  mkdir -p ./frontend/build
  touch ./frontend/build/index.html
  cd backend && go test ./...
}

build_frontend() {
  echo "building frontend"
  root=$(pwd)

  if [ -n "$FUSION_VERSION" ]; then
    version="$FUSION_VERSION"
  else
    # Try to get version relative to the last git tag.
    if git describe --tags --abbrev=0 >/dev/null 2>&1; then
      version=$(git describe --tags --abbrev=0)
    else
      # If repo has no tags, just use the latest commit hash.
      version=$(git rev-parse --short HEAD)
    fi
  fi
  echo "Using fusion version string: ${version}"

  cd ./frontend
  pnpm i
  VITE_FUSION_VERSION="$version" pnpm run build
  cd "$root"
}

build_backend() {
  target_os=${1:-$(go env GOOS)}
  target_arch=${2:-$(go env GOARCH)}
  echo "building backend for OS: ${target_os}, Arch: ${target_arch}"
  cd backend
  CGO_ENABLED=0 GOOS=${target_os} GOARCH=${target_arch} go build \
    -ldflags '-extldflags "-static"' \
    -o ../build/fusion \
    ./cmd/fusion
}

release() {
  echo "building release artifacts"
  rm -rf ./dist
  mkdir -p ./dist

  platforms="linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64"
  root=$(pwd)

  for platform in $platforms; do
    os=$(echo "$platform" | cut -d/ -f1)
    arch=$(echo "$platform" | cut -d/ -f2)
    echo "--- building ${os}/${arch} ---"

    bin_name="fusion"
    if [ "$os" = "windows" ]; then
      bin_name="fusion.exe"
    fi

    cd "$root/backend"
    CGO_ENABLED=0 GOOS=${os} GOARCH=${arch} go build \
      -ldflags '-extldflags "-static"' \
      -o "$root/dist/${bin_name}" \
      ./cmd/fusion
    cd "$root"

    # Create zip archive
    archive="fusion_${os}_${arch}.zip"
    zip -j "./dist/${archive}" "./dist/${bin_name}" LICENSE* README* || \
      zip -j "./dist/${archive}" "./dist/${bin_name}"
    rm "./dist/${bin_name}"
  done

  # Generate checksums
  cd ./dist
  sha256sum *.zip > checksums.txt
  cd "$root"

  echo "release artifacts:"
  ls -lh ./dist/
}


build() {
  test_go
  build_frontend
  build_backend
}

case $1 in
"test")
  test_go
  ;;
"build-frontend")
  build_frontend
  ;;
"build-backend")
  build_backend "$2" "$3"
  ;;
"build")
  build
  ;;
"release")
  release
  ;;
esac
