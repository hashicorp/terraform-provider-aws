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
		"DataSource": {
			"basic":                         testAccAppSyncDataSource_basic,
			"description":                   testAccAppSyncDataSource_description,
			"DynamoDB_region":               testAccAppSyncDataSource_DynamoDB_region,
			"DynamoDB_useCallerCredentials": testAccAppSyncDataSource_DynamoDB_useCallerCredentials,
			"ElasticSearch_region":          testAccAppSyncDataSource_ElasticSearch_region,
			"HTTP_endpoint":                 testAccAppSyncDataSource_HTTP_endpoint,
			"type":                          testAccAppSyncDataSource_type,
			"Type_dynamoDB":                 testAccAppSyncDataSource_Type_dynamoDB,
			"Type_elasticSearch":            testAccAppSyncDataSource_Type_elasticSearch,
			"Type_http":                     testAccAppSyncDataSource_Type_http,
			"Type_lambda":                   testAccAppSyncDataSource_Type_lambda,
			"Type_none":                     testAccAppSyncDataSource_Type_none,
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
