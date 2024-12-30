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
  cd ./frontend
  npm i
  npm run build
  cd $root
}

build_backend() {
  echo "building backend"
  go build \
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
