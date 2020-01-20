// Package S015 defines an Analyzer that checks for
// Schema that attribute names contain only lowercase
// alphanumerics and underscores
package S015

import (
	"go/ast"
	"regexp"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/bflad/tfproviderlint/passes/commentignore"
	"github.com/bflad/tfproviderlint/passes/schemamap"
)

const Doc = `check for Schema that attribute names are valid

The S015 analyzer reports cases of schemas which the attribute name
includes characters outside lowercase alphanumerics and underscores,
which will fail provider schema validation.`

const analyzerName = "S015"

var Analyzer = &analysis.Analyzer{
	Name: analyzerName,
	Doc:  Doc,
	Requires: []*analysis.Analyzer{
		schemamap.Analyzer,
		commentignore.Analyzer,
	},
	Run: run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	ignorer := pass.ResultOf[commentignore.Analyzer].(*commentignore.Ignorer)
	schemamaps := pass.ResultOf[schemamap.Analyzer].([]*ast.CompositeLit)

	attributeNameRegex := regexp.MustCompile(`^[a-z0-9_]+$`)

	for _, smap := range schemamaps {
		if ignorer.ShouldIgnore(analyzerName, smap) {
			continue
		}

		for _, attributeName := range schemamap.GetSchemaAttributeNames(smap) {
			switch t := attributeName.(type) {
			default:
				continue
			case *ast.BasicLit:
				value := strings.Trim(t.Value, `"`)

				if !attributeNameRegex.MatchString(value) {
					pass.Reportf(t.Pos(), "%s: schema attribute names should only be lowercase alphanumeric characters or underscores", analyzerName)
				}
			}
		}
	}

	return nil, nil
}
