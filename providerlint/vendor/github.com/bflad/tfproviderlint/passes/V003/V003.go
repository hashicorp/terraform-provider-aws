package V003

import (
	"github.com/bflad/tfproviderlint/helper/analysisutils"
	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/validation"
	"github.com/bflad/tfproviderlint/passes/helper/validation/iprangecallexpr"
	"github.com/bflad/tfproviderlint/passes/helper/validation/iprangeselectorexpr"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

var Analyzer = analysisutils.DeprecatedEmptyCallExprWithReplacementSelectorExprAnalyzer(
	"V003",
	iprangecallexpr.Analyzer,
	iprangeselectorexpr.Analyzer,
	validation.PackagePath,
	validation.FuncNameIPRange,
	validation.PackagePath,
	validation.FuncNameIsIPv4Range,
)
