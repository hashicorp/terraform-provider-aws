// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package resourcedatapartialselectorexpr

import (
	"github.com/bflad/tfproviderlint/helper/analysisutils"
	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/schema"
)

var Analyzer = analysisutils.ReceiverMethodSelectorExprAnalyzer(
	"resourcedatapartialselectorexpr",
	schema.IsReceiverMethod,
	schema.PackagePath,
	schema.TypeNameResourceData,
	"Partial",
)
