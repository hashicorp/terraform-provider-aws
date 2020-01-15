package terraformtype

import (
	"go/ast"
	"go/types"

	"github.com/bflad/tfproviderlint/helper/astutils"
)

const (
	FuncNameComposeAggregateTestCheckFunc = `ComposeAggregateTestCheckFunc`
	FuncNameComposeTestCheckFunc          = `ComposeTestCheckFunc`
	FuncNameNonRetryableError             = `NonRetryableError`
	FuncNameParallelTest                  = `ParallelTest`
	FuncNameRetryableError                = `RetryableError`
	FuncNameTest                          = `Test`
	FuncNameTestCheckNoResourceAttr       = `TestCheckNoResourceAttr`
	FuncNameTestCheckOutput               = `TestCheckOutput`
	FuncNameTestCheckResourceAttr         = `TestCheckResourceAttr`
	FuncNameTestCheckResourceAttrPair     = `TestCheckResourceAttrPair`
	FuncNameTestCheckResourceAttrPtr      = `TestCheckResourceAttrPtr`
	FuncNameTestCheckResourceAttrSet      = `TestCheckResourceAttrSet`
	FuncNameTestMatchResourceAttr         = `TestMatchResourceAttr`

	PackagePathHelperResource = `github.com/hashicorp/terraform-plugin-sdk/helper/resource`
)

// IsHelperResourceFunc returns if the function call is in the resource package
func IsHelperResourceFunc(e ast.Expr, info *types.Info, funcName string) bool {
	return astutils.IsPackageFunc(e, info, PackagePathHelperResource, funcName)
}

// IsHelperResourceNamedType returns if the type name matches and is from the helper/resource package
func IsHelperResourceNamedType(t *types.Named, typeName string) bool {
	return astutils.IsPackageNamedType(t, PackagePathHelperResource, typeName)
}
