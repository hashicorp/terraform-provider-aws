// Package XR002 defines an Analyzer that checks for
// Resource that should implement Importer
package XR002

import (
	"golang.org/x/tools/go/analysis"

	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/schema"
	"github.com/bflad/tfproviderlint/passes/commentignore"
	"github.com/bflad/tfproviderlint/passes/helper/schema/resourceinfo"
)

const Doc = `check for Resource that should implement Importer

The XR002 analyzer reports missing usage of Importer in resources.`

const analyzerName = "XR002"

var Analyzer = &analysis.Analyzer{
	Name: analyzerName,
	Doc:  Doc,
	Requires: []*analysis.Analyzer{
		resourceinfo.Analyzer,
		commentignore.Analyzer,
	},
	Run: run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	ignorer := pass.ResultOf[commentignore.Analyzer].(*commentignore.Ignorer)
	resources := pass.ResultOf[resourceinfo.Analyzer].([]*schema.ResourceInfo)
	for _, resource := range resources {
		if ignorer.ShouldIgnore(analyzerName, resource.AstCompositeLit) {
			continue
		}

		if !resource.IsResource() {
			continue
		}

		if resource.DeclaresField(schema.ResourceFieldImporter) {
			continue
		}

		pass.Reportf(resource.AstCompositeLit.Pos(), "%s: resource should include Importer implementation", analyzerName)
	}

	return nil, nil
}
