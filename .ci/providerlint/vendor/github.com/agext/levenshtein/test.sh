# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

set -ev

if [[ "$1" == "goveralls" ]]; then
	echo "Testing with goveralls..."
	go get github.com/mattn/goveralls
	$HOME/gopath/bin/goveralls -service=travis-ci
else
	echo "Testing with go test..."
	go test -v ./...
fi
