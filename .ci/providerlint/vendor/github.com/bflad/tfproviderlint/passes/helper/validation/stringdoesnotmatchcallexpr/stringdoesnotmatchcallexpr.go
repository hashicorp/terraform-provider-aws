// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package stringdoesnotmatchcallexpr

import (
	"github.com/bflad/tfproviderlint/helper/analysisutils"
	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/validation"
)

var Analyzer = analysisutils.FunctionCallExprAnalyzer(
	"stringdoesnotmatchcallexpr",
	validation.IsFunc,
	validation.PackagePath,
	validation.FuncNameStringDoesNotMatch,
)
