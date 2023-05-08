package version

import (
	_ "embed"

	"github.com/hashicorp/go-version"
)

//go:embed VERSION
var rawVersion string

// Version returns the version of the library
//
// Note: This is only exposed as public function/package
// due to hard-coded constraints in the release tooling.
// In general downstream should not implement version-specific
// logic and rely on this function to be present in future releases.
func Version() *version.Version {
	return version.Must(version.NewVersion(rawVersion))
}
