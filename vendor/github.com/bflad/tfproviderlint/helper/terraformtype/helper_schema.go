package terraformtype

import (
	"go/ast"
	"go/types"
)

const (
	FuncNameImportStatePassthrough = `ImportStatePassthrough`
	FuncNameNoop                   = `Noop`

	PackagePathHelperSchema = `github.com/hashicorp/terraform-plugin-sdk/helper/schema`
)

// IsHelperResourceFunc returns if the function call is in the helper/schema package
func IsHelperSchemaFunc(e ast.Expr, info *types.Info, funcName string) bool {
	return isPackageFunc(e, info, PackagePathHelperSchema, funcName)
}

// IsHelperSchemaNamedType returns if the type name matches and is from the helper/schema package
func IsHelperSchemaNamedType(t *types.Named, typeName string) bool {
	return isPackageNamedType(t, PackagePathHelperSchema, typeName)
}
