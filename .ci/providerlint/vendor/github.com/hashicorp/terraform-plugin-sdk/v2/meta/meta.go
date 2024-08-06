// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// The meta package provides a location to set the release version
// and any other relevant metadata for the SDK.
//
// This package should not import any other SDK packages.
package meta

import (
	"fmt"

	version "github.com/hashicorp/go-version"
)

// The main version number that is being run at the moment.
//
// Deprecated: Use Go standard library [runtime/debug] package build information
// instead.
var SDKVersion = "2.34.0"

// A pre-release marker for the version. If this is "" (empty string)
// then it means that it is a final release. Otherwise, this is a pre-release
// such as "dev" (in development), "beta", "rc1", etc.
//
// Deprecated: Use Go standard library [runtime/debug] package build information
// instead.
var SDKPrerelease = ""

// SemVer is an instance of version.Version. This has the secondary
// benefit of verifying during tests and init time that our version is a
// proper semantic version, which should always be the case.
var SemVer *version.Version

func init() {
	SemVer = version.Must(version.NewVersion(SDKVersion))
}

// VersionString returns the complete version string, including prerelease
//
// Deprecated: Use Go standard library [runtime/debug] package build information
// instead.
func SDKVersionString() string {
	if SDKPrerelease != "" {
		return fmt.Sprintf("%s-%s", SDKVersion, SDKPrerelease)
	}
	return SDKVersion
}
