// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

//go:build tools
// +build tools

package main

import (
	_ "github.com/YakDriver/copyplop"
	_ "github.com/YakDriver/tfproviderdocs"
	_ "github.com/client9/misspell/cmd/misspell"
	_ "github.com/golangci/golangci-lint/v2/cmd/golangci-lint"
	_ "github.com/hashicorp/go-changelog/cmd/changelog-build"
	_ "github.com/katbyte/terrafmt"
	_ "github.com/pavius/impi/cmd/impi"
	_ "github.com/rhysd/actionlint/cmd/actionlint"
	_ "github.com/terraform-linters/tflint"
	_ "golang.org/x/tools/cmd/stringer"
	_ "mvdan.cc/gofumpt"
)
