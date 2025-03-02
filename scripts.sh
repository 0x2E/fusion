#!/bin/sh

# Exit on first failure.
set -e

test_go() {
  echo "testing"
  # make some files for embed
  mkdir -p ./frontend/build
  touch ./frontend/build/index.html
  go test ./...
}

build_frontend() {
  echo "building frontend"
  mkdir -p ./build
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
  npm i
  VITE_FUSION_VERSION="$version" npm run build
  cd $root
}

build_backend() {
  echo "building backend"
  CGO_ENABLED=0 go build \
    -ldflags '-extldflags "-static"' \
    -o ./build/fusion \
    ./cmd/server/*
}

build() {
  test_go
  build_frontend
  build_backend
}

dev() {
  go run \
    -ldflags '-extldflags "-static"' \
    ./cmd/server
}

case $1 in
"test")
  test_go
  ;;
"dev")
  dev
  ;;
"build-frontend")
  build_frontend
  ;;
"build-backend")
  build_backend
  ;;
"build")
  build
  ;;
esac
