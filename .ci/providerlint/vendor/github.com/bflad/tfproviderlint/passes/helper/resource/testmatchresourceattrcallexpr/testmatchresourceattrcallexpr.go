// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testmatchresourceattrcallexpr

import (
	"github.com/bflad/tfproviderlint/helper/analysisutils"
	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/resource"
)

var Analyzer = analysisutils.FunctionCallExprAnalyzer(
	"testmatchresourceattrcallexpr",
	resource.IsFunc,
	resource.PackagePath,
	resource.FuncNameTestMatchResourceAttr,
)
