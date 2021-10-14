package cur_test

import (
	"testing"
	"time"
)

// Serialize to limit API rate-limit exceeded errors.
func TestAccCUR_serial(t *testing.T) {
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

	for group, m := range testCases {
		m := m
		t.Run(group, func(t *testing.T) {
			for name, tc := range m {
				tc := tc
				t.Run(name, func(t *testing.T) {
					tc(t)
					// Explicitly sleep between tests.
					time.Sleep(5 * time.Second)
				})
			}
		})
	}
}
