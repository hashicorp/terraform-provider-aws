#!/usr/bin/env bash
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0


# Check gofmt
echo "==> Checking that code complies with gofmt requirements..."
gofmt_files=$(find . -name '*.go' | grep -v vendor | xargs gofmt -l -s)
if [[ -n ${gofmt_files} ]]; then
    echo 'gofmt needs running on the following files:'
    echo "${gofmt_files}"
    echo "You can use the command: \`make fmt\` to reformat code."
    exit 1
fi

exit 0
