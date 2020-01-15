package terraformtype

import (
	"go/types"
)

const (
	TypeNameResourceData = `ResourceData`
)

// IsHelperSchemaTypeResourceData returns if the type is ResourceData from the helper/schema package
func IsHelperSchemaTypeResourceData(t types.Type) bool {
	switch t := t.(type) {
	case *types.Named:
		return IsHelperSchemaNamedType(t, TypeNameResourceData)
	case *types.Pointer:
		return IsHelperSchemaTypeResourceData(t.Elem())
	default:
		return false
	}
}
