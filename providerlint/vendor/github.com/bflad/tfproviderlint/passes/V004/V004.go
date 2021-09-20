package V004

import (
	"github.com/bflad/tfproviderlint/helper/analysisutils"
	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/validation"
	"github.com/bflad/tfproviderlint/passes/helper/validation/singleipcallexpr"
	"github.com/bflad/tfproviderlint/passes/helper/validation/singleipselectorexpr"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

var Analyzer = analysisutils.DeprecatedEmptyCallExprWithReplacementSelectorExprAnalyzer(
	"V004",
	singleipcallexpr.Analyzer,
	singleipselectorexpr.Analyzer,
	validation.PackagePath,
	validation.FuncNameSingleIP,
	validation.PackagePath,
	validation.FuncNameIsIPAddress,
)
