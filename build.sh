#!/bin/sh
set -e
VERSION="0.1.0-$(date +%Y%m%d%H%M%S)"
go build -ldflags "-X 'github.com/kno-ai/kno/internal.Version=$VERSION'" -o ./kno ./cmd/kno
echo "built kno $VERSION"
