// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package timesleepcallexpr

import (
	"github.com/bflad/tfproviderlint/helper/analysisutils"
)

var Analyzer = analysisutils.StdlibFunctionCallExprAnalyzer(
	"timesleepcallexpr",
	"time",
	"Sleep",
)
