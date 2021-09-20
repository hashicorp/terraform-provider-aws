package osexeccommandcontextcallexpr

import (
	"github.com/bflad/tfproviderlint/helper/analysisutils"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

var Analyzer = analysisutils.StdlibFunctionCallExprAnalyzer(
	"osexeccommandcontextcallexpr",
	"os/exec",
	"CommandContext",
)
