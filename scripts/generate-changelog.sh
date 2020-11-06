#!/bin/bash

__dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
__parent="$(dirname "$__dir")"
CHANGELOG_FILE_NAME="CHANGELOG.md"

TARGET_SHA=$(git rev-parse HEAD)
PREVIOUS_RELEASE_SHA=$(git rev-list -n 1 $(git describe --abbrev=0 --match='v*.*.*' --tags))

if [ $TARGET_SHA == $PREVIOUS_RELEASE_SHA ]; then
  echo "Nothing to do"
  exit 0
fi

CHANGELOG=$($(go env GOPATH)/bin/changelog-build -this-release $TARGET_SHA \
                      -last-release $PREVIOUS_RELEASE_SHA \
                      -git-dir $__parent \
                      -entries-dir .changelog \
                      -changelog-template $__dir/changelog.tmpl \
                      -note-template $__dir/release-note.tmpl)

if [ -z "$CHANGELOG" ]
then
    echo "No changelog generated."
    exit 0
fi

mkdir tmp

sed -n -e "1{/## /p;}" $__parent/$CHANGELOG_FILE_NAME > tmp/$CHANGELOG_FILE_NAME
echo "$CHANGELOG" >> tmp/$CHANGELOG_FILE_NAME
echo >> tmp/$CHANGELOG_FILE_NAME
sed -n -e "/## $(git describe --abbrev=0 --match='v*.*.*' --tags | tr -d v)/,\$p" $__parent/$CHANGELOG_FILE_NAME >> tmp/$CHANGELOG_FILE_NAME

cp tmp/$CHANGELOG_FILE_NAME $CHANGELOG_FILE_NAME

rm -rf tmp