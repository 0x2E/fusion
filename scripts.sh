#!/bin/sh

build() {
	root=$(pwd)
	rm -r ./build
	echo "building frontend"
	cd ./frontend && npm i && npm run build && cp -R ./build/ $root/build/frontend || exit 1
	cd $root || exit 1
	echo "building backend"
	cp .env build/
	go build -o ./build/fusion ./cmd/server/* || exit 1
}

gen() {
	go generate ./...
}

test_go() {
	gen
	go test ./...
}

case $1 in
"test")
	test_go
	;;
"gen")
	gen
	;;
"build")
	build
	;;
esac
