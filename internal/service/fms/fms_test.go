package fms_test

import (
	"testing"
)

func TestAccFMS_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"AdminAccount": {
			"basic": testAccAdminAccount_basic,
		},
		"Policy": {
			"basic":                  TestAccFMSPolicy_basic,
			"cloudfrontDistribution": TestAccFMSPolicy_cloudFrontDistribution,
			"includeMap":             TestAccFMSPolicy_includeMap,
			"update":                 TestAccFMSPolicy_update,
			"resourceTags":           TestAccFMSPolicy_resourceTags,
			"tags":                   TestAccFMSPolicy_tags,
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
