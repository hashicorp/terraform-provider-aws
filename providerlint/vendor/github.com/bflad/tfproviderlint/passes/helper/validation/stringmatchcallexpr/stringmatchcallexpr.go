package stringmatchcallexpr

import (
	"github.com/bflad/tfproviderlint/helper/analysisutils"
	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

var Analyzer = analysisutils.FunctionCallExprAnalyzer(
	"stringmatchcallexpr",
	validation.IsFunc,
	validation.PackagePath,
	validation.FuncNameStringMatch,
)
