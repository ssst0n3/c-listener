#!/bin/bash
set -ex
version=$(cat VERSION)
# go get github.com/mitchellh/gox
cd "$(dirname "$(readlink -m "$0")")"
rm -rf bin/release/$version
mkdir -p bin/release/$version
cd bin/release/$version
CGO_ENABLED=0 gox -cgo=0 -osarch="linux/amd64" -osarch="linux/arm64" -ldflags "${LDFLAGS}" github.com/ssst0n3/fd-listener/cmd/fd-listener
cd -
upx bin/release/$version/*