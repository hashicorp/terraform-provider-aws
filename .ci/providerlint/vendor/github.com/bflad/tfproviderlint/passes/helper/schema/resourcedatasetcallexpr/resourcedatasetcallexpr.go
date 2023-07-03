// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package resourcedatasetcallexpr

import (
	"github.com/bflad/tfproviderlint/helper/analysisutils"
	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/schema"
)

var Analyzer = analysisutils.ReceiverMethodCallExprAnalyzer(
	"resourcedatasetcallexpr",
	schema.IsReceiverMethod,
	schema.PackagePath,
	schema.TypeNameResourceData,
	"Set",
)
