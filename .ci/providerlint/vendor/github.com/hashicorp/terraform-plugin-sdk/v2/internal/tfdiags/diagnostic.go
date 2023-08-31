// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tfdiags

type Diagnostic interface {
	Severity() Severity
	Description() Description
}

type Severity rune

// This code was previously generated with a go:generate directive calling:
// go run golang.org/x/tools/cmd/stringer -type=Severity
// However, it is now considered frozen and the tooling dependency has been
// removed. The String method can be manually updated if necessary.

const (
	Error   Severity = 'E'
	Warning Severity = 'W'
)

type Description struct {
	Summary string
	Detail  string
}
