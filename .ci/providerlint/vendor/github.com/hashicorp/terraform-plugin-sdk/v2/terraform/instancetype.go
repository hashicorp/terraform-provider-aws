package terraform

// This code was previously generated with a go:generate directive calling:
// go run golang.org/x/tools/cmd/stringer -type=instanceType instancetype.go
// However, it is now considered frozen and the tooling dependency has been
// removed. The String method can be manually updated if necessary.

// instanceType is an enum of the various types of instances store in the State
type instanceType int

const (
	typeInvalid instanceType = iota
	typePrimary
	typeTainted
	typeDeposed
)
