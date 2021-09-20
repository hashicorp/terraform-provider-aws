package R007

import (
	"github.com/bflad/tfproviderlint/helper/analysisutils"
	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/schema"
	"github.com/bflad/tfproviderlint/passes/helper/schema/resourcedatapartialcallexpr"
	"github.com/bflad/tfproviderlint/passes/helper/schema/resourcedatapartialselectorexpr"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

var Analyzer = analysisutils.DeprecatedReceiverMethodSelectorExprAnalyzer(
	"R007",
	resourcedatapartialcallexpr.Analyzer,
	resourcedatapartialselectorexpr.Analyzer,
	schema.PackagePath,
	schema.TypeNameResourceData,
	"Partial",
)
