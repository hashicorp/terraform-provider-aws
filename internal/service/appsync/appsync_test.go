package appsync_test

import (
	"testing"
)

func TestAccAppSync_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"APIKey": {
			"basic":       testAccAppSyncAPIKey_basic,
			"description": testAccAppSyncAPIKey_description,
			"expires":     testAccAppSyncAPIKey_expires,
		},
		"Function": {
			"basic":                   testAccAppSyncFunction_basic,
			"disappears":              testAccAppSyncFunction_disappears,
			"description":             testAccAppSyncFunction_description,
			"responseMappingTemplate": testAccAppSyncFunction_responseMappingTemplate,
		},
		"Resolver": {
			"basic":             testAccAppSyncResolver_basic,
			"disappears":        testAccAppSyncResolver_disappears,
			"dataSource":        testAccAppSyncResolver_dataSource,
			"DataSource_lambda": testAccAppSyncResolver_DataSource_lambda,
			"requestTemplate":   testAccAppSyncResolver_requestTemplate,
			"responseTemplate":  testAccAppSyncResolver_responseTemplate,
			"multipleResolvers": testAccAppSyncResolver_multipleResolvers,
			"pipeline":          testAccAppSyncResolver_pipeline,
			"caching":           testAccAppSyncResolver_caching,
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
