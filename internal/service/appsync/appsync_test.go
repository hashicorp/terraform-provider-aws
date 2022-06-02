package appsync_test

import (
	"os"
	"testing"
)

func TestAccAppSync_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"APIKey": {
			"basic":       testAccAPIKey_basic,
			"description": testAccAPIKey_description,
			"expires":     testAccAPIKey_expires,
		},
		"DataSource": {
			"basic":                         testAccDataSource_basic,
			"description":                   testAccDataSource_description,
			"DynamoDB_region":               testAccDataSource_DynamoDB_region,
			"DynamoDB_useCallerCredentials": testAccDataSource_DynamoDB_useCallerCredentials,
			"HTTP_endpoint":                 testAccDataSource_HTTP_endpoint,
			"type":                          testAccDataSource_type,
			"Type_dynamoDB":                 testAccDataSource_Type_dynamoDB,
			"Type_http":                     testAccDataSource_Type_http,
			"Type_http_auth":                testAccDataSource_Type_httpAuth,
			"Type_lambda":                   testAccDataSource_Type_lambda,
			"Type_none":                     testAccDataSource_Type_none,
			"Type_rdbms":                    testAccDataSource_Type_relationalDatabase,
			"Type_rdbms_options":            testAccDataSource_Type_relationalDatabaseWithOptions,
		},
		"GraphQLAPI": {
			"basic":                     testAccGraphQLAPI_basic,
			"disappears":                testAccGraphQLAPI_disappears,
			"schema":                    testAccGraphQLAPI_schema,
			"authenticationType":        testAccGraphQLAPI_authenticationType,
			"AuthenticationType_apiKey": testAccGraphQLAPI_AuthenticationType_apiKey,
			"AuthenticationType_awsIAM": testAccGraphQLAPI_AuthenticationType_iam,
			"AuthenticationType_amazonCognitoUserPools": testAccGraphQLAPI_AuthenticationType_amazonCognitoUserPools,
			"AuthenticationType_openIDConnect":          testAccGraphQLAPI_AuthenticationType_openIDConnect,
			"AuthenticationType_awsLambda":              testAccGraphQLAPI_AuthenticationType_lambda,
			"log":                                       testAccGraphQLAPI_log,
			"Log_fieldLogLevel":                         testAccGraphQLAPI_Log_fieldLogLevel,
			"Log_excludeVerboseContent":                 testAccGraphQLAPI_Log_excludeVerboseContent,
			"OpenIDConnect_authTTL":                     testAccGraphQLAPI_OpenIDConnect_authTTL,
			"OpenIDConnect_clientID":                    testAccGraphQLAPI_OpenIDConnect_clientID,
			"OpenIDConnect_iatTTL":                      testAccGraphQLAPI_OpenIDConnect_iatTTL,
			"OpenIDConnect_issuer":                      testAccGraphQLAPI_OpenIDConnect_issuer,
			"name":                                      testAccGraphQLAPI_name,
			"UserPool_awsRegion":                        testAccGraphQLAPI_UserPool_region,
			"UserPool_defaultAction":                    testAccGraphQLAPI_UserPool_defaultAction,
			"LambdaAuthorizerConfig_authorizerUri":      testAccGraphQLAPI_LambdaAuthorizerConfig_authorizerURI,
			"LambdaAuthorizerConfig_identityValidationExpression": testAccGraphQLAPI_LambdaAuthorizerConfig_identityValidationExpression,
			"LambdaAuthorizerConfig_authorizerResultTtlInSeconds": testAccGraphQLAPI_LambdaAuthorizerConfig_authorizerResultTTLInSeconds,
			"tags":                                      testAccGraphQLAPI_tags,
			"AdditionalAuthentication_apiKey":           testAccGraphQLAPI_AdditionalAuthentication_apiKey,
			"AdditionalAuthentication_awsIAM":           testAccGraphQLAPI_AdditionalAuthentication_iam,
			"AdditionalAuthentication_cognitoUserPools": testAccGraphQLAPI_AdditionalAuthentication_cognitoUserPools,
			"AdditionalAuthentication_openIDConnect":    testAccGraphQLAPI_AdditionalAuthentication_openIDConnect,
			"AdditionalAuthentication_awsLambda":        testAccGraphQLAPI_AdditionalAuthentication_lambda,
			"AdditionalAuthentication_multiple":         testAccGraphQLAPI_AdditionalAuthentication_multiple,
			"xrayEnabled":                               testAccGraphQLAPI_xrayEnabled,
		},
		"Function": {
			"basic":                   testAccFunction_basic,
			"disappears":              testAccFunction_disappears,
			"description":             testAccFunction_description,
			"responseMappingTemplate": testAccFunction_responseMappingTemplate,
			"sync":                    testAccFunction_syncConfig,
		},
		"Resolver": {
			"basic":             testAccResolver_basic,
			"disappears":        testAccResolver_disappears,
			"dataSource":        testAccResolver_dataSource,
			"DataSource_lambda": testAccResolver_DataSource_lambda,
			"requestTemplate":   testAccResolver_requestTemplate,
			"responseTemplate":  testAccResolver_responseTemplate,
			"multipleResolvers": testAccResolver_multipleResolvers,
			"pipeline":          testAccResolver_pipeline,
			"caching":           testAccResolver_caching,
			"sync":              testAccResolver_syncConfig,
		},
		"ApiCache": {
			"basic":      testAccAPICache_basic,
			"disappears": testAccAPICache_disappears,
		},
		"DomainName": {
			"basic":       testAccDomainName_basic,
			"disappears":  testAccDomainName_disappears,
			"description": testAccDomainName_description,
		},
		"DomainNameAssociation": {
			"basic":      testAccDomainNameAPIAssociation_basic,
			"disappears": testAccDomainNameAPIAssociation_disappears,
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

func getCertDomain(t *testing.T) string {
	value := os.Getenv("AWS_APPSYNC_DOMAIN_NAME_CERTIFICATE_DOMAIN")
	if value == "" {
		t.Skip(
			"Environment variable AWS_APPSYNC_DOMAIN_NAME_CERTIFICATE_DOMAIN is not set. " +
				"This environment variable must be set to any non-empty value " +
				"to enable the test.")
	}

	return value
}
