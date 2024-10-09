// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cur_test

import (
	"testing"
	"time"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

// Serialize to limit API rate-limit exceeded errors.
func TestAccCUR_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"ReportDefinition": {
			acctest.CtBasic:         testAccReportDefinition_basic,
			"tags":                  testAccReportDefinition_tags,
			acctest.CtDisappears:    testAccReportDefinition_disappears,
			"textOrCsv":             testAccReportDefinition_textOrCSV,
			"parquet":               testAccReportDefinition_parquet,
			"athena":                testAccReportDefinition_athena,
			"refresh":               testAccReportDefinition_refresh,
			"overwrite":             testAccReportDefinition_overwrite,
			"DataSource_basic":      testAccReportDefinitionDataSource_basic,
			"DataSource_additional": testAccReportDefinitionDataSource_additional,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 5*time.Second)
}
