// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fmtsprintfcallexpr

import (
	"github.com/bflad/tfproviderlint/helper/analysisutils"
)

var Analyzer = analysisutils.StdlibFunctionCallExprAnalyzer(
	"fmtsprintfcallexpr",
	"fmt",
	"Sprintf",
)
