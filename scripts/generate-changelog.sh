#!/bin/bash

__dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
__parent="$(dirname "$__dir")"

PREVIOUS_RELEASE=$(git rev-list -n 1 $(git describe --abbrev=0 --match='v*.*.*' --tags))

go get github.com/hashicorp/go-changelog/cmd/changelog-build

CHANGELOG=$($(go env GOPATH)/bin/changelog-build -this-release $(git rev-parse HEAD) \
                      -last-release $PREVIOUS_RELEASE \
                      -git-dir $__parent \
                      -entries-dir .changelog \
                      -changelog-template $__dir/changelog.tmpl \
                      -note-template $__dir/release-note.tmpl)

mkdir tmp

sed -n -e "1{/## /p;}" $__parent/CHANGELOG.md > tmp/CHANGELOG.md
echo "$CHANGELOG" >> tmp/CHANGELOG.md
echo >> tmp/CHANGELOG.md
sed -n -e "/## $(git describe --abbrev=0 --match='v*.*.*' --tags | tr -d v)/,\$p" $__parent/CHANGELOG.md >> tmp/CHANGELOG.md

cp tmp/CHANGELOG.md CHANGELOG.md

rm -rf tmp