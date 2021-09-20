package testmatchresourceattrcallexpr

import (
	"github.com/bflad/tfproviderlint/helper/analysisutils"
	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

var Analyzer = analysisutils.FunctionCallExprAnalyzer(
	"testmatchresourceattrcallexpr",
	resource.IsFunc,
	resource.PackagePath,
	resource.FuncNameTestMatchResourceAttr,
)
