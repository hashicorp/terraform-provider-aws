package terraformtype

import (
	"go/ast"
	"go/types"

	"github.com/bflad/tfproviderlint/helper/astutils"
)

const (
	FuncNameImportStatePassthrough = `ImportStatePassthrough`
	FuncNameNoop                   = `Noop`

	PackagePathHelperSchema = `github.com/hashicorp/terraform-plugin-sdk/helper/schema`
)

// IsHelperResourceFunc returns if the function call is in the helper/schema package
func IsHelperSchemaFunc(e ast.Expr, info *types.Info, funcName string) bool {
	return astutils.IsPackageFunc(e, info, PackagePathHelperSchema, funcName)
}

// IsHelperSchemaNamedType returns if the type name matches and is from the helper/schema package
func IsHelperSchemaNamedType(t *types.Named, typeName string) bool {
	return astutils.IsPackageNamedType(t, PackagePathHelperSchema, typeName)
}

func IsHelperSchemaReceiverMethod(e ast.Expr, info *types.Info, receiverName string, methodName string) bool {
	return astutils.IsPackageReceiverMethod(e, info, PackagePathHelperSchema, receiverName, methodName)
}
