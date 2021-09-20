package V002

import (
	"github.com/bflad/tfproviderlint/helper/analysisutils"
	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/validation"
	"github.com/bflad/tfproviderlint/passes/helper/validation/cidrnetworkselectorexpr"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

var Analyzer = analysisutils.DeprecatedWithReplacementSelectorExprAnalyzer(
	"V002",
	cidrnetworkselectorexpr.Analyzer,
	validation.PackagePath,
	validation.FuncNameCIDRNetwork,
	validation.PackagePath,
	validation.FuncNameIsCIDRNetwork,
)
