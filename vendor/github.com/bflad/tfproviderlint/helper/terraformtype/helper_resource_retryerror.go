package terraformtype

import (
	"go/ast"
	"go/types"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

const (
	RetryErrorFieldErr       = `Err`
	RetryErrorFieldRetryable = `Retryable`

	TypeNameRetryError = `RetryError`
)

// HelperResourceRetryErrorInfo represents all gathered RetryError data for easier access
type HelperResourceRetryErrorInfo struct {
	AstCompositeLit *ast.CompositeLit
	Fields          map[string]*ast.KeyValueExpr
	RetryError      *resource.RetryError
	TypesInfo       *types.Info
}

// NewHelperResourceRetryErrorInfo instantiates a HelperResourceRetryErrorInfo
func NewHelperResourceRetryErrorInfo(cl *ast.CompositeLit, info *types.Info) *HelperResourceRetryErrorInfo {
	result := &HelperResourceRetryErrorInfo{
		AstCompositeLit: cl,
		Fields:          astCompositeLitFields(cl),
		RetryError:      &resource.RetryError{},
		TypesInfo:       info,
	}

	return result
}

// DeclaresField returns true if the field name is present in the AST
func (info *HelperResourceRetryErrorInfo) DeclaresField(fieldName string) bool {
	return info.Fields[fieldName] != nil
}

// IsHelperResourceTypeRetryError returns if the type is RetryError from the helper/resource package
func IsHelperResourceTypeRetryError(t types.Type) bool {
	switch t := t.(type) {
	case *types.Named:
		return IsHelperResourceNamedType(t, TypeNameRetryError)
	case *types.Pointer:
		return IsHelperResourceTypeRetryError(t.Elem())
	default:
		return false
	}
}
