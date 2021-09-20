package osexeccommandcallexpr

import (
	"github.com/bflad/tfproviderlint/helper/analysisutils"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

var Analyzer = analysisutils.StdlibFunctionCallExprAnalyzer(
	"osexeccommandcallexpr",
	"os/exec",
	"Command",
)
