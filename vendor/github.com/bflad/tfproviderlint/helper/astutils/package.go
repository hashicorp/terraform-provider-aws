package astutils

import (
	"fmt"
	"go/ast"
	"go/types"
	"regexp"
	"strings"
)

// IsModulePackageReceiverMethod returns true if the module and package suffix (for vendoring), receiver name, and method name match
//
// This function automatically handles Go module versioning in import paths.
// To explicitly check an import path, use IsPackageReceiverMethod instead.
func IsModulePackageReceiverMethod(e ast.Expr, info *types.Info, module string, packageSuffix string, receiverName string, methodName string) bool {
	switch e := e.(type) {
	case *ast.SelectorExpr:
		if e.Sel.Name != methodName {
			return false
		}

		return IsModulePackageType(info.TypeOf(e.X), module, packageSuffix, receiverName)
	}

	return false
}

// IsModulePackageFunc returns true if the function package suffix (for vendoring) and name matches
//
// This function automatically handles Go module versioning in import paths.
// To explicitly check an import path, use IsPackageFunc instead.
func IsModulePackageFunc(e ast.Expr, info *types.Info, module string, packageSuffix string, funcName string) bool {
	switch e := e.(type) {
	case *ast.SelectorExpr:
		if e.Sel.Name != funcName {
			return false
		}

		switch x := e.X.(type) {
		case *ast.Ident:
			return isModulePackagePath(module, packageSuffix, info.ObjectOf(x).(*types.PkgName).Imported().Path())
		}
	case *ast.StarExpr:
		return IsModulePackageFunc(e.X, info, module, packageSuffix, funcName)
	}

	return false
}

// IsModulePackageFunctionFieldListType returns true if the function parameter package suffix (for vendoring) and name matches
//
// This function automatically handles Go module versioning in import paths.
// To explicitly check an import path, use IsPackageFunctionFieldListType instead.
func IsModulePackageFunctionFieldListType(e ast.Expr, info *types.Info, module string, packageSuffix string, typeName string) bool {
	switch e := e.(type) {
	case *ast.SelectorExpr:
		if e.Sel.Name != typeName {
			return false
		}

		switch x := e.X.(type) {
		case *ast.Ident:
			return isModulePackagePath(module, packageSuffix, info.ObjectOf(x).(*types.PkgName).Imported().Path())
		}
	case *ast.StarExpr:
		return IsModulePackageFunctionFieldListType(e.X, info, module, packageSuffix, typeName)
	}

	return false
}

// IsModulePackageNamedType returns if the type name matches and is from the package suffix
//
// This function automatically handles Go module versioning in import paths.
// To explicitly check an import path, use IsPackageNamedType instead.
func IsModulePackageNamedType(t *types.Named, module string, packageSuffix string, typeName string) bool {
	if t.Obj().Name() != typeName {
		return false
	}

	return isModulePackagePath(module, packageSuffix, t.Obj().Pkg().Path())
}

// IsModulePackageType returns true if the type name can be matched and is from the package suffix
//
// This function automatically handles Go module versioning in import paths.
// To explicitly check an import path, use IsPackageType instead.
func IsModulePackageType(t types.Type, module string, packageSuffix string, typeName string) bool {
	switch t := t.(type) {
	case *types.Named:
		return IsModulePackageNamedType(t, module, packageSuffix, typeName)
	case *types.Pointer:
		return IsModulePackageType(t.Elem(), module, packageSuffix, typeName)
	}

	return false
}

// IsPackageReceiverMethod returns true if the package suffix (for vendoring), receiver name, and method name match
//
// This function checks an explicit import path. To allow any Go module version
// in the import path, use IsModulePackageReceiverMethod instead.
func IsPackageReceiverMethod(e ast.Expr, info *types.Info, packageSuffix string, receiverName, methodName string) bool {
	switch e := e.(type) {
	case *ast.SelectorExpr:
		if e.Sel.Name != methodName {
			return false
		}

		return IsPackageType(info.TypeOf(e.X), packageSuffix, receiverName)
	}

	return false
}

// IsPackageFunc returns true if the function package suffix (for vendoring) and name matches
//
// This function checks an explicit import path. To allow any Go module version
// in the import path, use IsModulePackageFunc instead.
func IsPackageFunc(e ast.Expr, info *types.Info, packageSuffix string, funcName string) bool {
	switch e := e.(type) {
	case *ast.SelectorExpr:
		if e.Sel.Name != funcName {
			return false
		}

		switch x := e.X.(type) {
		case *ast.Ident:
			return strings.HasSuffix(info.ObjectOf(x).(*types.PkgName).Imported().Path(), packageSuffix)
		}
	case *ast.StarExpr:
		return IsPackageFunc(e.X, info, packageSuffix, funcName)
	}

	return false
}

// IsPackageFunctionFieldListType returns true if the function parameter package suffix (for vendoring) and name matches
//
// This function checks an explicit import path. To allow any Go module version
// in the import path, use IsModuleFunctionFieldListType instead.
func IsPackageFunctionFieldListType(e ast.Expr, info *types.Info, packageSuffix string, typeName string) bool {
	switch e := e.(type) {
	case *ast.SelectorExpr:
		if e.Sel.Name != typeName {
			return false
		}

		switch x := e.X.(type) {
		case *ast.Ident:
			return strings.HasSuffix(info.ObjectOf(x).(*types.PkgName).Imported().Path(), packageSuffix)
		}
	case *ast.StarExpr:
		return IsPackageFunctionFieldListType(e.X, info, packageSuffix, typeName)
	}

	return false
}

// IsPackageNamedType returns if the type name matches and is from the package suffix
//
// This function checks an explicit import path. To allow any Go module version
// in the import path, use IsModulePackageNamedType instead.
func IsPackageNamedType(t *types.Named, packageSuffix string, typeName string) bool {
	if t.Obj().Name() != typeName {
		return false
	}

	// HasSuffix here due to vendoring
	return strings.HasSuffix(t.Obj().Pkg().Path(), packageSuffix)
}

// IsPackageType returns true if the type name can be matched and is from the package suffix
//
// This function checks an explicit import path. To allow any Go module version
// in the import path, use IsModulePackageType instead.
func IsPackageType(t types.Type, packageSuffix string, typeName string) bool {
	switch t := t.(type) {
	case *types.Named:
		return IsPackageNamedType(t, packageSuffix, typeName)
	case *types.Pointer:
		return IsPackageType(t.Elem(), packageSuffix, typeName)
	}

	return false
}

func isModulePackagePath(module string, packageSuffix string, path string) bool {
	// Only check end of path due to vendoring
	r := regexp.MustCompile(fmt.Sprintf("%s(/v[1-9][0-9]*)?/%s$", module, packageSuffix))
	return r.MatchString(path)
}
