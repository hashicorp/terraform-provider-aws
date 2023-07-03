// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package resourcedatasetpartialcallexpr

import (
	"github.com/bflad/tfproviderlint/helper/analysisutils"
	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/schema"
)

var Analyzer = analysisutils.ReceiverMethodCallExprAnalyzer(
	"resourcedatasetpartialcallexpr",
	schema.IsReceiverMethod,
	schema.PackagePath,
	schema.TypeNameResourceData,
	"SetPartial",
)
