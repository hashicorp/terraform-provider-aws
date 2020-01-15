package terraformtype

import (
	"go/ast"
	"go/types"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

const (
	ResourceFieldCreate             = `Create`
	ResourceFieldCustomizeDiff      = `CustomizeDiff`
	ResourceFieldDelete             = `Delete`
	ResourceFieldDeprecationMessage = `DeprecationMessage`
	ResourceFieldExists             = `Exists`
	ResourceFieldImporter           = `Importer`
	ResourceFieldMigrateState       = `MigrateState`
	ResourceFieldRead               = `Read`
	ResourceFieldSchema             = `Schema`
	ResourceFieldSchemaVersion      = `SchemaVersion`
	ResourceFieldStateUpgraders     = `StateUpgraders`
	ResourceFieldTimeouts           = `Timeouts`
	ResourceFieldUpdate             = `Update`

	TypeNameResource = `Resource`
)

// HelperSchemaResourceInfo represents all gathered Resource data for easier access
type HelperSchemaResourceInfo struct {
	AstCompositeLit *ast.CompositeLit
	Fields          map[string]*ast.KeyValueExpr
	Resource        *schema.Resource
	TypesInfo       *types.Info
}

// NewHelperSchemaResourceInfo instantiates a HelperSchemaResourceInfo
func NewHelperSchemaResourceInfo(cl *ast.CompositeLit, info *types.Info) *HelperSchemaResourceInfo {
	result := &HelperSchemaResourceInfo{
		AstCompositeLit: cl,
		Fields:          astCompositeLitFields(cl),
		Resource:        &schema.Resource{},
		TypesInfo:       info,
	}

	return result
}

// DeclaresField returns true if the field name is present in the AST
func (info *HelperSchemaResourceInfo) DeclaresField(fieldName string) bool {
	return info.Fields[fieldName] != nil
}

// IsHelperSchemaTypeResource returns if the type is Resource from the helper/schema package
func IsHelperSchemaTypeResource(t types.Type) bool {
	switch t := t.(type) {
	case *types.Named:
		return IsHelperSchemaNamedType(t, TypeNameResource)
	case *types.Pointer:
		return IsHelperSchemaTypeResource(t.Elem())
	default:
		return false
	}
}
