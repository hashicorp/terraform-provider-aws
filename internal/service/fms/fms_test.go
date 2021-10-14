package fms_test

import (
	"testing"
)

func TestAccAWSFms_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"AdminAccount": {
			"basic": testAccAdminAccount_basic,
		},
		"Policy": {
			"basic":                  testAccPolicy_basic,
			"cloudfrontDistribution": testAccPolicy_cloudFrontDistribution,
			"includeMap":             testAccPolicy_includeMap,
			"update":                 testAccPolicy_update,
			"tags":                   testAccPolicy_tags,
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
