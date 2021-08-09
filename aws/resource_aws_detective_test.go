package aws

import (
	"testing"
)

func TestAccAWSDetective_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"Graph": {
			"basic": testAccAwsDetectiveGraph_basic,
			"tags":  testAccAwsDetectiveGraph_WithTags,
		},
	}

	for group, m := range testCases {
		m := m
		t.Run(group, func(t *testing.T) {
			for name, tc := range m {
				tc := tc
				t.Run(name, func(t *testing.T) {
					tc(t)
				})
			}
		})
	}
}
