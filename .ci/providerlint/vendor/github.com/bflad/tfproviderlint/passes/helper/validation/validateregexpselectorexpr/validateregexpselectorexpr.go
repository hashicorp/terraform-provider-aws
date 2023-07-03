// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package validateregexpselectorexpr

import (
	"github.com/bflad/tfproviderlint/helper/analysisutils"
	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/validation"
)

var Analyzer = analysisutils.SelectorExprAnalyzer(
	"validateregexpselectorexpr",
	validation.IsFunc,
	validation.PackagePath,
	validation.FuncNameValidateRegexp,
)
