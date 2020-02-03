package terraformtype

import (
	"go/ast"
	"go/token"
	"go/types"
)

// HelperSchemaSchemaValidateFuncInfo represents all gathered SchemaValidateFunc data for easier access
type HelperSchemaSchemaValidateFuncInfo struct {
	AstFuncDecl *ast.FuncDecl
	AstFuncLit  *ast.FuncLit
	Body        *ast.BlockStmt
	Node        ast.Node
	Pos         token.Pos
	Type        *ast.FuncType
	TypesInfo   *types.Info
}

// NewHelperSchemaSchemaValidateFuncInfo instantiates a HelperSchemaSchemaValidateFuncInfo
func NewHelperSchemaSchemaValidateFuncInfo(funcDecl *ast.FuncDecl, funcLit *ast.FuncLit, info *types.Info) *HelperSchemaSchemaValidateFuncInfo {
	result := &HelperSchemaSchemaValidateFuncInfo{
		AstFuncDecl: funcDecl,
		AstFuncLit:  funcLit,
		TypesInfo:   info,
	}

	if funcDecl != nil {
		result.Body = funcDecl.Body
		result.Node = funcDecl
		result.Pos = funcDecl.Pos()
		result.Type = funcDecl.Type
	} else if funcLit != nil {
		result.Body = funcLit.Body
		result.Node = funcLit
		result.Pos = funcLit.Pos()
		result.Type = funcLit.Type
	}

	return result
}
