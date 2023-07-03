// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package S035

import (
	"github.com/bflad/tfproviderlint/helper/analysisutils"
	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/schema"
)

var Analyzer = analysisutils.SchemaAttributeReferencesAnalyzer("S035", schema.SchemaFieldAtLeastOneOf)
