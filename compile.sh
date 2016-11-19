#!/bin/bash

set -e

cd /go/src/swarm-prune
trash
go vet
go build -tags netgo
cp swarm-prune /go/bin
