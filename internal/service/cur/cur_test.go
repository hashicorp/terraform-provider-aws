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
			"basic":      testAccReportDefinition_basic,
			"disappears": testAccReportDefinition_disappears,
			"textOrCsv":  testAccReportDefinition_textOrCSV,
			"parquet":    testAccReportDefinition_parquet,
			"athena":     testAccReportDefinition_athena,
			"refresh":    testAccReportDefinition_refresh,
			"overwrite":  testAccReportDefinition_overwrite,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 5*time.Second)
}
