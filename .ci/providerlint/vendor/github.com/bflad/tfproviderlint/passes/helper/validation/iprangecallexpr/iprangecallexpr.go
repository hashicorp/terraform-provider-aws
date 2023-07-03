// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iprangecallexpr

import (
	"github.com/bflad/tfproviderlint/helper/analysisutils"
	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/validation"
)

var Analyzer = analysisutils.FunctionCallExprAnalyzer(
	"iprangecallexpr",
	validation.IsFunc,
	validation.PackagePath,
	validation.FuncNameIPRange,
)
