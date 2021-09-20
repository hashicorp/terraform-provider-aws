package osexeccommandselectorexpr

import (
	"github.com/bflad/tfproviderlint/helper/analysisutils"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

var Analyzer = analysisutils.StdlibFunctionSelectorExprAnalyzer(
	"osexeccommandselectorexpr",
	"os/exec",
	"Command",
)
