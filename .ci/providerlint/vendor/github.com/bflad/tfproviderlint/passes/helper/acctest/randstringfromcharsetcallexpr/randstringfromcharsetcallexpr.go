// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package randstringfromcharsetcallexpr

import (
	"github.com/bflad/tfproviderlint/helper/analysisutils"
	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/acctest"
)

var Analyzer = analysisutils.FunctionCallExprAnalyzer(
	"randstringfromcharsetcallexpr",
	acctest.IsFunc,
	acctest.PackagePath,
	acctest.FuncNameRandStringFromCharSet,
)
