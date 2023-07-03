// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package S037

import (
	"github.com/bflad/tfproviderlint/helper/analysisutils"
	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/schema"
)

var Analyzer = analysisutils.SchemaAttributeReferencesAnalyzer("S037", schema.SchemaFieldExactlyOneOf)
