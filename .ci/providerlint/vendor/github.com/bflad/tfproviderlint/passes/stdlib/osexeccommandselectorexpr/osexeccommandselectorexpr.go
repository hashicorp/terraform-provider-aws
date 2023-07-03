// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package osexeccommandselectorexpr

import (
	"github.com/bflad/tfproviderlint/helper/analysisutils"
)

var Analyzer = analysisutils.StdlibFunctionSelectorExprAnalyzer(
	"osexeccommandselectorexpr",
	"os/exec",
	"Command",
)
