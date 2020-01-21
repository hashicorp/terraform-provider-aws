package terraformtype

import (
	"go/ast"
	"go/types"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

const (
	TestStepFieldCheck                     = `Check`
	TestStepFieldConfig                    = `Config`
	TestStepFieldDestroy                   = `Destroy`
	TestStepFieldExpectError               = `ExpectError`
	TestStepFieldExpectNonEmptyPlan        = `ExpectNonEmptyPlan`
	TestStepFieldImportState               = `ImportState`
	TestStepFieldImportStateId             = `ImportStateId`
	TestStepFieldImportStateIdFunc         = `ImportStateIdFunc`
	TestStepFieldImportStateIdPrefix       = `ImportStateIdPrefix`
	TestStepFieldImportStateCheck          = `ImportStateCheck`
	TestStepFieldImportStateVerify         = `ImportStateVerify`
	TestStepFieldImportStateVerifyIgnore   = `ImportStateVerifyIgnore`
	TestStepFieldPlanOnly                  = `PlanOnly`
	TestStepFieldPreConfig                 = `PreConfig`
	TestStepFieldPreventDiskCleanup        = `PreventDiskCleanup`
	TestStepFieldPreventPostDestroyRefresh = `PreventPostDestroyRefresh`
	TestStepFieldResourceName              = `ResourceName`
	TestStepFieldSkipFunc                  = `SkipFunc`
	TestStepFieldTaint                     = `Taint`

	TypeNameTestStep = `TestStep`
)

// HelperResourceTestStepInfo represents all gathered TestStep data for easier access
type HelperResourceTestStepInfo struct {
	AstCompositeLit *ast.CompositeLit
	Fields          map[string]*ast.KeyValueExpr
	TestStep        *resource.TestStep
	TypesInfo       *types.Info
}

// NewHelperResourceTestStepInfo instantiates a HelperResourceTestStepInfo
func NewHelperResourceTestStepInfo(cl *ast.CompositeLit, info *types.Info) *HelperResourceTestStepInfo {
	result := &HelperResourceTestStepInfo{
		AstCompositeLit: cl,
		Fields:          astCompositeLitFields(cl),
		TestStep:        &resource.TestStep{},
		TypesInfo:       info,
	}

	return result
}

// DeclaresField returns true if the field name is present in the AST
func (info *HelperResourceTestStepInfo) DeclaresField(fieldName string) bool {
	return info.Fields[fieldName] != nil
}

// IsHelperResourceTypeTestStep returns if the type is TestStep from the helper/schema package
func IsHelperResourceTypeTestStep(t types.Type) bool {
	switch t := t.(type) {
	case *types.Named:
		return IsHelperResourceNamedType(t, TypeNameTestStep)
	case *types.Pointer:
		return IsHelperResourceTypeTestStep(t.Elem())
	default:
		return false
	}
}
