// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package osexeccommandcontextcallexpr

import (
	"github.com/bflad/tfproviderlint/helper/analysisutils"
)

var Analyzer = analysisutils.StdlibFunctionCallExprAnalyzer(
	"osexeccommandcontextcallexpr",
	"os/exec",
	"CommandContext",
)
