#!/usr/bin/env bash
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0


set -e
echo "" > coverage.txt

for d in $(go list ./... | grep -v vendor); do
    go test -race -coverprofile=profile.out -covermode=atomic "$d"
    if [ -f profile.out ]; then
        cat profile.out >> coverage.txt
        rm profile.out
    fi
done
