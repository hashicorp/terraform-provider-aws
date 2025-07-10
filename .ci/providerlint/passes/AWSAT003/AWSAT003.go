// Package AWSAT003 defines an Analyzer that checks for
// hardcoded regions
package AWSAT003

import (
	"go/ast"
	"go/token"
	"regexp"
	"strings"

	"github.com/bflad/tfproviderlint/passes/commentignore"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const Doc = `check for hardcoded regions

The AWSAT003 analyzer reports hardcoded regions. Testing in non-standard
partitions with hardcoded regions (and AZs) will cause the tests to fail. 
`

const analyzerName = "AWSAT003"

var Analyzer = &analysis.Analyzer{
	Name: analyzerName,
	Doc:  Doc,
	Requires: []*analysis.Analyzer{
		commentignore.Analyzer,
		inspect.Analyzer,
	},
	Run: run,
}

func run(pass *analysis.Pass) (any, error) {
	ignorer := pass.ResultOf[commentignore.Analyzer].(*commentignore.Ignorer)
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.BasicLit)(nil),
	}

	partitions := endpoints.DefaultPartitions()
	var regions []string

	for _, p := range partitions {
		for id := range p.Regions() {
			regions = append(regions, id)
		}
	}

	re := regexp.MustCompile(strings.Join(regions, "|"))
	inspect.Preorder(nodeFilter, func(n ast.Node) {
		x := n.(*ast.BasicLit)

		if ignorer.ShouldIgnore(analyzerName, x) {
			return
		}

		if x.Kind != token.STRING {
			return
		}

		if !re.MatchString(x.Value) {
			return
		}

		pass.Reportf(x.ValuePos, "%s: regions should not be hardcoded, use aws_region and aws_availability_zones data sources instead", analyzerName)
	})
	return nil, nil
}
