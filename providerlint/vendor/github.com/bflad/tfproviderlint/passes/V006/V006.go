package V006

import (
	"github.com/bflad/tfproviderlint/helper/analysisutils"
	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/validation"
	"github.com/bflad/tfproviderlint/passes/helper/validation/validatelistuniquestringsselectorexpr"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

var Analyzer = analysisutils.DeprecatedWithReplacementSelectorExprAnalyzer(
	"V006",
	validatelistuniquestringsselectorexpr.Analyzer,
	validation.PackagePath,
	validation.FuncNameValidateListUniqueStrings,
	validation.PackagePath,
	validation.FuncNameListOfUniqueStrings,
)
