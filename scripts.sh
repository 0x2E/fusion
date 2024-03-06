#!/bin/sh

build() {
	root=$(pwd)
	rm -r ./build
	mkdir -p ./build/frontend
	echo "building backend"
	cp .env build/
	go build -o ./build/server ./cmd/server/* || exit 1
	echo "building frontend"
	cd ./frontend && npm run build && cp -R ./build/ $root/build/frontend || exit 1
	cd $root || exit 1
}

case $1 in
"test")
	go test ./...
	;;
"gen")
	go generate ./...
	;;
"build")
	build
	;;
esac
