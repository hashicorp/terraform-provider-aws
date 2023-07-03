// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testcheckresourceattrcallexpr

import (
	"github.com/bflad/tfproviderlint/helper/analysisutils"
	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/resource"
)

var Analyzer = analysisutils.FunctionCallExprAnalyzer(
	"testcheckresourceattrcallexpr",
	resource.IsFunc,
	resource.PackagePath,
	resource.FuncNameTestCheckResourceAttr,
)
