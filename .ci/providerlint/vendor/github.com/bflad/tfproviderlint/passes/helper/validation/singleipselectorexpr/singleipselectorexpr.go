// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package singleipselectorexpr

import (
	"github.com/bflad/tfproviderlint/helper/analysisutils"
	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/validation"
)

var Analyzer = analysisutils.SelectorExprAnalyzer(
	"singleipselectorexpr",
	validation.IsFunc,
	validation.PackagePath,
	validation.FuncNameSingleIP,
)
