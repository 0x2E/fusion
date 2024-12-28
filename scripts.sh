#!/bin/sh

# Exit on first failure.
set -e

gen() {
  go generate ./...
}

test_go() {
  echo "testing"
  gen
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
  gen
  go run \
    -ldflags '-extldflags "-static"' \
    ./cmd/server
}

case $1 in
"test")
  test_go
  ;;
"gen")
  gen
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
