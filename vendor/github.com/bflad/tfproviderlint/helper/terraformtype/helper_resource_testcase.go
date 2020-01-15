package terraformtype

import (
	"go/ast"
	"go/types"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

const (
	TestCaseFieldCheckDestroy              = `CheckDestroy`
	TestCaseFieldIDRefreshName             = `IDRefreshName`
	TestCaseFieldIDRefreshIgnore           = `IDRefreshIgnore`
	TestCaseFieldIsUnitTest                = `IsUnitTest`
	TestCaseFieldPreCheck                  = `PreCheck`
	TestCaseFieldPreventPostDestroyRefresh = `PreventPostDestroyRefresh`
	TestCaseFieldProviders                 = `Providers`
	TestCaseFieldProviderFactories         = `ProviderFactories`
	TestCaseFieldSteps                     = `Steps`

	TypeNameTestCase = `TestCase`
)

// HelperResourceTestCaseInfo represents all gathered TestCase data for easier access
type HelperResourceTestCaseInfo struct {
	AstCompositeLit *ast.CompositeLit
	Fields          map[string]*ast.KeyValueExpr
	TestCase        *resource.TestCase
	TypesInfo       *types.Info
}

// NewHelperResourceTestCaseInfo instantiates a HelperResourceTestCaseInfo
func NewHelperResourceTestCaseInfo(cl *ast.CompositeLit, info *types.Info) *HelperResourceTestCaseInfo {
	result := &HelperResourceTestCaseInfo{
		AstCompositeLit: cl,
		Fields:          astCompositeLitFields(cl),
		TestCase:        &resource.TestCase{},
		TypesInfo:       info,
	}

	return result
}

// DeclaresField returns true if the field name is present in the AST
func (info *HelperResourceTestCaseInfo) DeclaresField(fieldName string) bool {
	return info.Fields[fieldName] != nil
}

// IsHelperResourceTypeTestCase returns if the type is TestCase from the helper/schema package
func IsHelperResourceTypeTestCase(t types.Type) bool {
	switch t := t.(type) {
	case *types.Named:
		return IsHelperResourceNamedType(t, TypeNameTestCase)
	case *types.Pointer:
		return IsHelperResourceTypeTestCase(t.Elem())
	default:
		return false
	}
}
