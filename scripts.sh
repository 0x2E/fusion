#!/bin/sh

# Exit on first failure.
set -e

gen() {
  go generate ./...
}

test_go() {
  gen
  # make some files for embed
  mkdir -p ./frontend/build
  touch ./frontend/build/index.html
  go test ./...
}

build() {
  echo "testing"
  gen
  test_go

  root=$(pwd)
  mkdir -p ./build
  echo "building frontend"
  cd ./frontend
  npm i
  npm run build
  cd $root
  echo "building backend"
  go build -o ./build/fusion ./cmd/server/*
}

dev() {
  gen
  go run ./cmd/server
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
"build")
  build
  ;;
esac
