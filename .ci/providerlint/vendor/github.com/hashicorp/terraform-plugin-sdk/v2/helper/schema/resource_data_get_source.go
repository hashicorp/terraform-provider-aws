package schema

// This code was previously generated with a go:generate directive calling:
// go run golang.org/x/tools/cmd/stringer -type=getSource resource_data_get_source.go
// However, it is now considered frozen and the tooling dependency has been
// removed. The String method can be manually updated if necessary.

// getSource represents the level we want to get for a value (internally).
// Any source less than or equal to the level will be loaded (whichever
// has a value first).
type getSource byte

const (
	getSourceState getSource = 1 << iota
	getSourceConfig
	getSourceDiff
	getSourceSet
	getSourceExact               // Only get from the _exact_ level
	getSourceLevelMask getSource = getSourceState | getSourceConfig | getSourceDiff | getSourceSet
)
