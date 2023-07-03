// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package validatejsonstringselectorexpr

import (
	"github.com/bflad/tfproviderlint/helper/analysisutils"
	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/validation"
)

var Analyzer = analysisutils.SelectorExprAnalyzer(
	"validatejsonstringselectorexpr",
	validation.IsFunc,
	validation.PackagePath,
	validation.FuncNameValidateJsonString,
)
