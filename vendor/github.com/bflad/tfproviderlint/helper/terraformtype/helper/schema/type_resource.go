package schema

import (
	"go/ast"
	"go/types"

	"github.com/bflad/tfproviderlint/helper/astutils"
	tfschema "github.com/hashicorp/terraform-plugin-sdk/helper/schema"
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

// ResourceInfo represents all gathered Resource data for easier access
type ResourceInfo struct {
	AstCompositeLit *ast.CompositeLit
	Fields          map[string]*ast.KeyValueExpr
	Resource        *tfschema.Resource
	TypesInfo       *types.Info
}

// NewResourceInfo instantiates a ResourceInfo
func NewResourceInfo(cl *ast.CompositeLit, info *types.Info) *ResourceInfo {
	result := &ResourceInfo{
		AstCompositeLit: cl,
		Fields:          astutils.CompositeLitFields(cl),
		Resource:        &tfschema.Resource{},
		TypesInfo:       info,
	}

	return result
}

// DeclaresField returns true if the field name is present in the AST
func (info *ResourceInfo) DeclaresField(fieldName string) bool {
	return info.Fields[fieldName] != nil
}

// IsDataSource returns true if the Resource type matches a Terraform Data Source declaration
func (info *ResourceInfo) IsDataSource() bool {
	if info.DeclaresField(ResourceFieldCreate) {
		return false
	}

	return info.DeclaresField(ResourceFieldRead)
}

// IsResource returns true if the Resource type matches a Terraform Resource declaration
func (info *ResourceInfo) IsResource() bool {
	return info.DeclaresField(ResourceFieldCreate)
}

// IsTypeResource returns if the type is Resource from the helper/schema package
func IsTypeResource(t types.Type) bool {
	switch t := t.(type) {
	case *types.Named:
		return IsNamedType(t, TypeNameResource)
	case *types.Pointer:
		return IsTypeResource(t.Elem())
	default:
		return false
	}
}
