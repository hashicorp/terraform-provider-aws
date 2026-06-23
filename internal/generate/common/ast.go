// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"iter"
	"os"
	"strings"

	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"golang.org/x/tools/go/packages"
)

type PackageFile struct {
	name string
	file *ast.File
}

func (file *PackageFile) Name() string {
	return file.name
}

func (file *PackageFile) File() *ast.File {
	return file.file
}

func (file *PackageFile) PackageName() string {
	return file.file.Name.Name
}

type Package struct {
	name  string
	files []*PackageFile
}

func (pkg *Package) Name() string {
	return pkg.name
}

func (pkg *Package) Files() []*PackageFile {
	return pkg.files
}

func LoadPackage(sourcePackage string) (*Package, error) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedTypes | packages.NeedSyntax,
	}
	pkgs, err := packages.Load(cfg, sourcePackage)
	if err != nil {
		return nil, fmt.Errorf("loading %s: %w", sourcePackage, err)
	}
	if len(pkgs) != 1 {
		return nil, fmt.Errorf("%d packages found", len(pkgs))
	}
	pkg := pkgs[0]

	return &Package{
		name: pkg.Name,
		files: tfslices.ApplyToAll(pkg.Syntax, func(file *ast.File) *PackageFile {
			return &PackageFile{
				file: file,
			}
		}),
	}, nil
}

func (pkg *Package) FindFunction(functionName string) *PackageFunction {
	for _, file := range pkg.Files() {
		if file != nil {
			for _, decl := range file.File().Decls {
				if funcDecl, ok := decl.(*ast.FuncDecl); ok {
					if funcDecl.Name.Name == functionName {
						return &PackageFunction{
							funcDecl: funcDecl,
						}
					}
				}
			}
		}
	}

	return nil
}

type PackageFunction struct {
	funcDecl *ast.FuncDecl
}

func (function *PackageFunction) Name() string {
	return function.funcDecl.Name.Name
}

func (function *PackageFunction) Params() []*ast.Field {
	return function.funcDecl.Type.Params.List
}

func (function *PackageFunction) Results() []*ast.Field {
	return function.funcDecl.Type.Results.List
}

// ScanDirectory scans a single directory and returns an iterator of Go sources files.
func ScanDirectory(path string) iter.Seq2[*PackageFile, error] {
	return func(yield func(*PackageFile, error) bool) {
		entries, err := os.ReadDir(path)
		if err != nil {
			yield(nil, err)
			return
		}

		fileSet := token.NewFileSet()

		for _, entry := range entries {
			// Skip directories, test files, and service_package_gen.go.
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") || entry.Name() == "service_package_gen.go" {
				continue
			}

			name := path + "/" + entry.Name()

			file, err := parser.ParseFile(fileSet, name, nil, parser.ParseComments)
			if err != nil {
				yield(nil, err)
				return
			}

			if !yield(&PackageFile{name: name, file: file}, nil) {
				return
			}
		}
	}
}
