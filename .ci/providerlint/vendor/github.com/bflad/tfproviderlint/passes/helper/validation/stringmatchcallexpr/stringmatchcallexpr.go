// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package stringmatchcallexpr

import (
	"github.com/bflad/tfproviderlint/helper/analysisutils"
	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/validation"
)

var Analyzer = analysisutils.FunctionCallExprAnalyzer(
	"stringmatchcallexpr",
	validation.IsFunc,
	validation.PackagePath,
	validation.FuncNameStringMatch,
)
