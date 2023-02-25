package schema

// This code was previously generated with a go:generate directive calling:
// go run golang.org/x/tools/cmd/stringer -type=ValueType valuetype.go
// However, it is now considered frozen and the tooling dependency has been
// removed. The String method can be manually updated if necessary.

// ValueType is an enum of the type that can be represented by a schema.
type ValueType int

const (
	TypeInvalid ValueType = iota
	TypeBool
	TypeInt
	TypeFloat
	TypeString
	TypeList
	TypeMap
	TypeSet
	typeObject
)

// NOTE: ValueType has more functions defined on it in schema.go. We can't
// put them here because we reference other files.
