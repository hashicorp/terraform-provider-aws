// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package version

import (
	_ "embed"
	"strings"

	"github.com/hashicorp/go-version"
)

//go:embed VERSION
var rawVersion string

// parsedVersion declared here ensures that invalid versions panic early, on import
var parsedVersion = version.Must(version.NewVersion(strings.TrimSpace(rawVersion)))

// Version returns the version of the library
//
// Note: This is only exposed as public function/package
// due to hard-coded constraints in the release tooling.
// In general downstream should not implement version-specific
// logic and rely on this function to be present in future releases.
func Version() *version.Version {
	return parsedVersion
}
