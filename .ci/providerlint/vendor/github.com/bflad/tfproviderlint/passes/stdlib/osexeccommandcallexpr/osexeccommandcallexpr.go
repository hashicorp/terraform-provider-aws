// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package osexeccommandcallexpr

import (
	"github.com/bflad/tfproviderlint/helper/analysisutils"
)

var Analyzer = analysisutils.StdlibFunctionCallExprAnalyzer(
	"osexeccommandcallexpr",
	"os/exec",
	"Command",
)
