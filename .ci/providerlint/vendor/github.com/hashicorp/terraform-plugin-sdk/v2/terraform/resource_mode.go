// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package terraform

// This code was previously generated with a go:generate directive calling:
// go run golang.org/x/tools/cmd/stringer -type=ResourceMode -output=resource_mode_string.go resource_mode.go
// However, it is now considered frozen and the tooling dependency has been
// removed. The String method can be manually updated if necessary.

// ResourceMode is deprecated, use addrs.ResourceMode instead.
// It has been preserved for backwards compatibility.
type ResourceMode int

const (
	ManagedResourceMode ResourceMode = iota
	DataResourceMode
)
