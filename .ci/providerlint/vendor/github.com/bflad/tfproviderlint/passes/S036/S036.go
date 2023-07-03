// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package S036

import (
	"github.com/bflad/tfproviderlint/helper/analysisutils"
	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/schema"
)

var Analyzer = analysisutils.SchemaAttributeReferencesAnalyzer("S036", schema.SchemaFieldConflictsWith)
