// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appsync_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccAppSync_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"APIKey": {
			acctest.CtBasic: testAccAPIKey_basic,
			"description":   testAccAPIKey_description,
			"expires":       testAccAPIKey_expires,
		},
		"DataSource": {
			acctest.CtBasic:                 testAccDataSource_basic,
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
			"Type_eventBridge":              testAccDataSource_Type_eventBridge,
		},
		"GraphQLAPI": {
			acctest.CtBasic:             testAccGraphQLAPI_basic,
			acctest.CtDisappears:        testAccGraphQLAPI_disappears,
			"tags":                      testAccGraphQLAPI_tags,
			"schema":                    testAccGraphQLAPI_schema,
			"authenticationType":        testAccGraphQLAPI_authenticationType,
			"AuthenticationType_apiKey": testAccGraphQLAPI_AuthenticationType_apiKey,
			"AuthenticationType_awsIAM": testAccGraphQLAPI_AuthenticationType_iam,
			"AuthenticationType_amazonCognitoUserPools":           testAccGraphQLAPI_AuthenticationType_amazonCognitoUserPools,
			"AuthenticationType_openIDConnect":                    testAccGraphQLAPI_AuthenticationType_openIDConnect,
			"AuthenticationType_awsLambda":                        testAccGraphQLAPI_AuthenticationType_lambda,
			"log":                                                 testAccGraphQLAPI_log,
			"Log_fieldLogLevel":                                   testAccGraphQLAPI_Log_fieldLogLevel,
			"Log_excludeVerboseContent":                           testAccGraphQLAPI_Log_excludeVerboseContent,
			"OpenIDConnect_authTTL":                               testAccGraphQLAPI_OpenIDConnect_authTTL,
			"OpenIDConnect_clientID":                              testAccGraphQLAPI_OpenIDConnect_clientID,
			"OpenIDConnect_iatTTL":                                testAccGraphQLAPI_OpenIDConnect_iatTTL,
			"OpenIDConnect_issuer":                                testAccGraphQLAPI_OpenIDConnect_issuer,
			acctest.CtName:                                        testAccGraphQLAPI_name,
			"UserPool_awsRegion":                                  testAccGraphQLAPI_UserPool_region,
			"UserPool_defaultAction":                              testAccGraphQLAPI_UserPool_defaultAction,
			"LambdaAuthorizerConfig_authorizerUri":                testAccGraphQLAPI_LambdaAuthorizerConfig_authorizerURI,
			"LambdaAuthorizerConfig_identityValidationExpression": testAccGraphQLAPI_LambdaAuthorizerConfig_identityValidationExpression,
			"LambdaAuthorizerConfig_authorizerResultTtlInSeconds": testAccGraphQLAPI_LambdaAuthorizerConfig_authorizerResultTTLInSeconds,
			"AdditionalAuthentication_apiKey":                     testAccGraphQLAPI_AdditionalAuthentication_apiKey,
			"AdditionalAuthentication_awsIAM":                     testAccGraphQLAPI_AdditionalAuthentication_iam,
			"AdditionalAuthentication_cognitoUserPools":           testAccGraphQLAPI_AdditionalAuthentication_cognitoUserPools,
			"AdditionalAuthentication_openIDConnect":              testAccGraphQLAPI_AdditionalAuthentication_openIDConnect,
			"AdditionalAuthentication_awsLambda":                  testAccGraphQLAPI_AdditionalAuthentication_lambda,
			"AdditionalAuthentication_multiple":                   testAccGraphQLAPI_AdditionalAuthentication_multiple,
			"xrayEnabled":                                         testAccGraphQLAPI_xrayEnabled,
			"visibility":                                          testAccGraphQLAPI_visibility,
			"introspectionConfig":                                 testAccGraphQLAPI_introspectionConfig,
			"queryDepthLimit":                                     testAccGraphQLAPI_queryDepthLimit,
			"resolverCountLimit":                                  testAccGraphQLAPI_resolverCountLimit,
		},
		"Function": {
			acctest.CtBasic:           testAccFunction_basic,
			"code":                    testAccFunction_code,
			acctest.CtDisappears:      testAccFunction_disappears,
			"description":             testAccFunction_description,
			"responseMappingTemplate": testAccFunction_responseMappingTemplate,
			"sync":                    testAccFunction_syncConfig,
		},
		"Resolver": {
			acctest.CtBasic:      testAccResolver_basic,
			"code":               testAccResolver_code,
			acctest.CtDisappears: testAccResolver_disappears,
			"dataSource":         testAccResolver_dataSource,
			"DataSource_lambda":  testAccResolver_DataSource_lambda,
			"requestTemplate":    testAccResolver_requestTemplate,
			"responseTemplate":   testAccResolver_responseTemplate,
			"multipleResolvers":  testAccResolver_multipleResolvers,
			"pipeline":           testAccResolver_pipeline,
			"caching":            testAccResolver_caching,
			"sync":               testAccResolver_syncConfig,
		},
		"ApiCache": {
			acctest.CtBasic:      testAccAPICache_basic,
			acctest.CtDisappears: testAccAPICache_disappears,
		},
		"Type": {
			acctest.CtBasic:      testAccType_basic,
			acctest.CtDisappears: testAccType_disappears,
		},
		"DomainName": {
			acctest.CtBasic:      testAccDomainName_basic,
			acctest.CtDisappears: testAccDomainName_disappears,
			"description":        testAccDomainName_description,
		},
		"DomainNameAssociation": {
			acctest.CtBasic:      testAccDomainNameAPIAssociation_basic,
			acctest.CtDisappears: testAccDomainNameAPIAssociation_disappears,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
