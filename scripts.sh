#!/bin/sh

build() {
	root=$(pwd)
	rm -r ./build
	echo "building frontend"
	cd ./frontend && npm i && npm run build || exit 1
	cd $root || exit 1
	echo "building backend"
	cp .env build/
	go build -o ./build/fusion ./cmd/server/* || exit 1
}

gen() {
	go generate ./...
}

test_go() {
	gen || exit 1
	# make some files for embed
	mkdir -p ./frontend/build && touch ./frontend/build/index.html || exit 1
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
