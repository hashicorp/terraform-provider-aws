package appsync_test

import (
	"os"
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
			"HTTP_endpoint":                 testAccAppSyncDataSource_HTTP_endpoint,
			"type":                          testAccAppSyncDataSource_type,
			"Type_dynamoDB":                 testAccAppSyncDataSource_Type_dynamoDB,
			"Type_http":                     testAccAppSyncDataSource_Type_http,
			"Type_http_auth":                testAccAppSyncDataSource_Type_http_auth,
			"Type_lambda":                   testAccAppSyncDataSource_Type_lambda,
			"Type_none":                     testAccAppSyncDataSource_Type_none,
			"Type_rdbms":                    testAccAppsyncDatasource_Type_RelationalDatabase,
			"Type_rdbms_options":            testAccAppsyncDatasource_Type_RelationalDatabaseWithOptions,
		},
		"GraphQLAPI": {
			"basic":                     testAccAppSyncGraphQLAPI_basic,
			"disappears":                testAccAppSyncGraphQLAPI_disappears,
			"schema":                    testAccAppSyncGraphQLAPI_schema,
			"authenticationType":        testAccAppSyncGraphQLAPI_authenticationType,
			"AuthenticationType_apiKey": testAccAppSyncGraphQLAPI_AuthenticationType_apiKey,
			"AuthenticationType_awsIAM": testAccAppSyncGraphQLAPI_AuthenticationType_awsIAM,
			"AuthenticationType_amazonCognitoUserPools": testAccAppSyncGraphQLAPI_AuthenticationType_amazonCognitoUserPools,
			"AuthenticationType_openIDConnect":          testAccAppSyncGraphQLAPI_AuthenticationType_openIDConnect,
			"AuthenticationType_awsLambda":              testAccAppSyncGraphQLAPI_AuthenticationType_awsLambda,
			"log":                                       testAccAppSyncGraphQLAPI_log,
			"Log_fieldLogLevel":                         testAccAppSyncGraphQLAPI_Log_fieldLogLevel,
			"Log_excludeVerboseContent":                 testAccAppSyncGraphQLAPI_Log_excludeVerboseContent,
			"OpenIDConnect_authTTL":                     testAccAppSyncGraphQLAPI_OpenIDConnect_authTTL,
			"OpenIDConnect_clientID":                    testAccAppSyncGraphQLAPI_OpenIDConnect_clientID,
			"OpenIDConnect_iatTTL":                      testAccAppSyncGraphQLAPI_OpenIDConnect_iatTTL,
			"OpenIDConnect_issuer":                      testAccAppSyncGraphQLAPI_OpenIDConnect_issuer,
			"name":                                      testAccAppSyncGraphQLAPI_name,
			"UserPool_awsRegion":                        testAccAppSyncGraphQLAPI_UserPool_awsRegion,
			"UserPool_defaultAction":                    testAccAppSyncGraphQLAPI_UserPool_defaultAction,
			"LambdaAuthorizerConfig_authorizerUri":      testAccAppSyncGraphQLAPI_LambdaAuthorizerConfig_authorizerUri,
			"LambdaAuthorizerConfig_identityValidationExpression": testAccAppSyncGraphQLAPI_LambdaAuthorizerConfig_identityValidationExpression,
			"LambdaAuthorizerConfig_authorizerResultTtlInSeconds": testAccAppSyncGraphQLAPI_LambdaAuthorizerConfig_authorizerResultTtlInSeconds,
			"tags":                                      testAccAppSyncGraphQLAPI_tags,
			"AdditionalAuthentication_apiKey":           testAccAppSyncGraphQLAPI_AdditionalAuthentication_apiKey,
			"AdditionalAuthentication_awsIAM":           testAccAppSyncGraphQLAPI_AdditionalAuthentication_awsIAM,
			"AdditionalAuthentication_cognitoUserPools": testAccAppSyncGraphQLAPI_AdditionalAuthentication_cognitoUserPools,
			"AdditionalAuthentication_openIDConnect":    testAccAppSyncGraphQLAPI_AdditionalAuthentication_openIDConnect,
			"AdditionalAuthentication_awsLambda":        testAccAppSyncGraphQLAPI_AdditionalAuthentication_awsLambda,
			"AdditionalAuthentication_multiple":         testAccAppSyncGraphQLAPI_AdditionalAuthentication_multiple,
			"xrayEnabled":                               testAccAppSyncGraphQLAPI_xrayEnabled,
		},
		"Function": {
			"basic":                   testAccAppSyncFunction_basic,
			"disappears":              testAccAppSyncFunction_disappears,
			"description":             testAccAppSyncFunction_description,
			"responseMappingTemplate": testAccAppSyncFunction_responseMappingTemplate,
			"sync":                    testAccAppSyncFunction_syncConfig,
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
			"sync":              testAccAppSyncResolver_syncConfig,
		},
		"ApiCache": {
			"basic":      testAccAppSyncApiCache_basic,
			"disappears": testAccAppSyncApiCache_disappears,
		},
		"DomainName": {
			"basic":       testAccAppSyncDomainName_basic,
			"disappears":  testAccAppSyncDomainName_disappears,
			"description": testAccAppSyncDomainName_description,
		},
		"DomainNameAssociation": {
			"basic":      testAccAppSyncDomainNameApiAssociation_basic,
			"disappears": testAccAppSyncDomainNameApiAssociation_disappears,
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

func getAppsyncCertDomain(t *testing.T) string {
	value := os.Getenv("AWS_APPSYNC_DOMAIN_NAME_CERTIFICATE_DOMAIN")
	if value == "" {
		t.Skip(
			"Environment variable AWS_APPSYNC_DOMAIN_NAME_CERTIFICATE_DOMAIN is not set. " +
				"This environment variable must be set to any non-empty value " +
				"to enable the test.")
	}

	return value
}
