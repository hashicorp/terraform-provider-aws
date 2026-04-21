// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package appsync_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appsync/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfappsync "github.com/hashicorp/terraform-provider-aws/internal/service/appsync"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAppSyncGraphQLAPI_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var api1 awstypes.GraphqlApi
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphQLAPIDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_authenticationType(rName, "API_KEY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, t, resourceName, &api1),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "appsync", regexache.MustCompile(`apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "API_KEY"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "log_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "user_pool_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "lambda_authorizer_config.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "uris.%"),
					resource.TestCheckResourceAttrSet(resourceName, "uris.GRAPHQL"),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrTags),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "xray_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "visibility", "GLOBAL"),
					resource.TestCheckResourceAttr(resourceName, "introspection_config", "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "query_depth_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "resolver_count_limit", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAppSyncGraphQLAPI_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var api1 awstypes.GraphqlApi
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphQLAPIDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_authenticationType(rName, "API_KEY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, t, resourceName, &api1),
					acctest.CheckSDKResourceDisappears(ctx, t, tfappsync.ResourceGraphQLAPI(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAppSyncGraphQLAPI_schema(t *testing.T) {
	ctx := acctest.Context(t)
	var api1, api2 awstypes.GraphqlApi
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphQLAPIDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_schema(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, t, resourceName, &api1),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "appsync", regexache.MustCompile(`apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "API_KEY"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "log_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "user_pool_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "lambda_authorizer_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "xray_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrSchema),
					resource.TestCheckResourceAttrSet(resourceName, "uris.%"),
					resource.TestCheckResourceAttrSet(resourceName, "uris.GRAPHQL"),
					testAccCheckGraphQLAPITypeExists(ctx, t, resourceName, "Post"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrSchema},
			},
			{
				Config: testAccGraphQLAPIConfig_schemaUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, t, resourceName, &api2),
					testAccCheckGraphQLAPITypeExists(ctx, t, resourceName, "PostV2"),
				),
			},
		},
	})
}

func TestAccAppSyncGraphQLAPI_authenticationType(t *testing.T) {
	ctx := acctest.Context(t)
	var api1, api2 awstypes.GraphqlApi
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphQLAPIDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_authenticationType(rName, "API_KEY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, t, resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "API_KEY"),
				),
			},
			{
				Config: testAccGraphQLAPIConfig_authenticationType(rName, "AWS_IAM"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, t, resourceName, &api2),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "AWS_IAM"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAppSyncGraphQLAPI_AuthenticationType_apiKey(t *testing.T) {
	ctx := acctest.Context(t)
	var api1 awstypes.GraphqlApi
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphQLAPIDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_authenticationType(rName, "API_KEY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, t, resourceName, &api1),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "appsync", regexache.MustCompile(`apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "API_KEY"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAppSyncGraphQLAPI_AuthenticationType_iam(t *testing.T) {
	ctx := acctest.Context(t)
	var api1 awstypes.GraphqlApi
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphQLAPIDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_authenticationType(rName, "AWS_IAM"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, t, resourceName, &api1),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "appsync", regexache.MustCompile(`apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "AWS_IAM"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAppSyncGraphQLAPI_AuthenticationType_amazonCognitoUserPools(t *testing.T) {
	ctx := acctest.Context(t)
	var api1 awstypes.GraphqlApi
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	cognitoUserPoolResourceName := "aws_cognito_user_pool.test"
	resourceName := "aws_appsync_graphql_api.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphQLAPIDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_userPoolDefaultAction(rName, "ALLOW"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, t, resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "AMAZON_COGNITO_USER_POOLS"),
					resource.TestCheckResourceAttr(resourceName, "user_pool_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "user_pool_config.0.aws_region", acctest.Region()),
					resource.TestCheckResourceAttr(resourceName, "user_pool_config.0.default_action", "ALLOW"),
					resource.TestCheckResourceAttrPair(resourceName, "user_pool_config.0.user_pool_id", cognitoUserPoolResourceName, names.AttrID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAppSyncGraphQLAPI_AuthenticationType_openIDConnect(t *testing.T) {
	ctx := acctest.Context(t)
	var api1 awstypes.GraphqlApi
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphQLAPIDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_openIDConnectIssuer(rName, "https://example.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, t, resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "OPENID_CONNECT"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.0.issuer", "https://example.com"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAppSyncGraphQLAPI_AuthenticationType_lambda(t *testing.T) {
	ctx := acctest.Context(t)
	var api1 awstypes.GraphqlApi
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"
	lambdaAuthorizerResourceName := "aws_lambda_function.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphQLAPIDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_lambdaAuthorizerAuthorizerURI(rName, "aws_lambda_function.test.arn"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, t, resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "AWS_LAMBDA"),
					resource.TestCheckResourceAttr(resourceName, "lambda_authorizer_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_authorizer_config.0.authorizer_uri", lambdaAuthorizerResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "lambda_authorizer_config.0.authorizer_result_ttl_in_seconds", strconv.Itoa(tfappsync.DefaultAuthorizerResultTTLInSeconds)),
					resource.TestCheckResourceAttr(resourceName, "lambda_authorizer_config.0.identity_validation_expression", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAppSyncGraphQLAPI_enhancedMetricsConfig(t *testing.T) {
	ctx := acctest.Context(t)
	var api1 awstypes.GraphqlApi
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphQLAPIDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_enhancedMetricsConfig(rName, "PER_DATA_SOURCE_METRICS", "ENABLED", "PER_RESOLVER_METRICS"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, t, resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "enhanced_metrics_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "enhanced_metrics_config.0.data_source_level_metrics_behavior", "PER_DATA_SOURCE_METRICS"),
					resource.TestCheckResourceAttr(resourceName, "enhanced_metrics_config.0.operation_level_metrics_config", "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "enhanced_metrics_config.0.resolver_level_metrics_behavior", "PER_RESOLVER_METRICS"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGraphQLAPIConfig_enhancedMetricsConfig(rName, "FULL_REQUEST_DATA_SOURCE_METRICS", "DISABLED", "FULL_REQUEST_RESOLVER_METRICS"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, t, resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "enhanced_metrics_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "enhanced_metrics_config.0.data_source_level_metrics_behavior", "FULL_REQUEST_DATA_SOURCE_METRICS"),
					resource.TestCheckResourceAttr(resourceName, "enhanced_metrics_config.0.operation_level_metrics_config", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "enhanced_metrics_config.0.resolver_level_metrics_behavior", "FULL_REQUEST_RESOLVER_METRICS"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAppSyncGraphQLAPI_log(t *testing.T) {
	ctx := acctest.Context(t)
	var api1 awstypes.GraphqlApi
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_appsync_graphql_api.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphQLAPIDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_logFieldLogLevel(rName, "ALL"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, t, resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "log_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "log_config.0.cloudwatch_logs_role_arn", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "log_config.0.field_log_level", "ALL"),
					resource.TestCheckResourceAttr(resourceName, "log_config.0.exclude_verbose_content", acctest.CtFalse),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAppSyncGraphQLAPI_Log_fieldLogLevel(t *testing.T) {
	ctx := acctest.Context(t)
	var api1, api2, api3 awstypes.GraphqlApi
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_appsync_graphql_api.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphQLAPIDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_logFieldLogLevel(rName, "ALL"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, t, resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "log_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "log_config.0.cloudwatch_logs_role_arn", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "log_config.0.field_log_level", "ALL"),
					resource.TestCheckResourceAttr(resourceName, "log_config.0.exclude_verbose_content", acctest.CtFalse),
				),
			},
			{
				Config: testAccGraphQLAPIConfig_logFieldLogLevel(rName, "ERROR"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, t, resourceName, &api2),
					resource.TestCheckResourceAttr(resourceName, "log_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "log_config.0.cloudwatch_logs_role_arn", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "log_config.0.field_log_level", "ERROR"),
					resource.TestCheckResourceAttr(resourceName, "log_config.0.exclude_verbose_content", acctest.CtFalse),
				),
			},
			{
				Config: testAccGraphQLAPIConfig_logFieldLogLevel(rName, "NONE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, t, resourceName, &api3),
					resource.TestCheckResourceAttr(resourceName, "log_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "log_config.0.cloudwatch_logs_role_arn", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "log_config.0.field_log_level", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "log_config.0.exclude_verbose_content", acctest.CtFalse),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAppSyncGraphQLAPI_Log_excludeVerboseContent(t *testing.T) {
	ctx := acctest.Context(t)
	var api1, api2 awstypes.GraphqlApi
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_appsync_graphql_api.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphQLAPIDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_logExcludeVerboseContent(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, t, resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "log_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "log_config.0.cloudwatch_logs_role_arn", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "log_config.0.field_log_level", "ALL"),
					resource.TestCheckResourceAttr(resourceName, "log_config.0.exclude_verbose_content", acctest.CtFalse),
				),
			},
			{
				Config: testAccGraphQLAPIConfig_logExcludeVerboseContent(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, t, resourceName, &api2),
					resource.TestCheckResourceAttr(resourceName, "log_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "log_config.0.cloudwatch_logs_role_arn", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "log_config.0.field_log_level", "ALL"),
					resource.TestCheckResourceAttr(resourceName, "log_config.0.exclude_verbose_content", acctest.CtTrue),
				),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAppSyncGraphQLAPI_OpenIDConnect_authTTL(t *testing.T) {
	ctx := acctest.Context(t)
	var api1, api2 awstypes.GraphqlApi
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphQLAPIDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_openIDConnectAuthTTL(rName, 1000),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, t, resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "OPENID_CONNECT"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.0.auth_ttl", "1000"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.0.issuer", "https://example.com"),
				),
			},
			{
				Config: testAccGraphQLAPIConfig_openIDConnectAuthTTL(rName, 2000),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, t, resourceName, &api2),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "OPENID_CONNECT"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.0.auth_ttl", "2000"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.0.issuer", "https://example.com"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAppSyncGraphQLAPI_OpenIDConnect_clientID(t *testing.T) {
	ctx := acctest.Context(t)
	var api1, api2 awstypes.GraphqlApi
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphQLAPIDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_openIDConnectClientID(rName, "ClientID1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, t, resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "OPENID_CONNECT"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.0.client_id", "ClientID1"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.0.issuer", "https://example.com"),
				),
			},
			{
				Config: testAccGraphQLAPIConfig_openIDConnectClientID(rName, "ClientID2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, t, resourceName, &api2),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "OPENID_CONNECT"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.0.client_id", "ClientID2"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.0.issuer", "https://example.com"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAppSyncGraphQLAPI_OpenIDConnect_iatTTL(t *testing.T) {
	ctx := acctest.Context(t)
	var api1, api2 awstypes.GraphqlApi
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphQLAPIDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_openIDConnectIatTTL(rName, 1000),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, t, resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "OPENID_CONNECT"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.0.iat_ttl", "1000"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.0.issuer", "https://example.com"),
				),
			},
			{
				Config: testAccGraphQLAPIConfig_openIDConnectIatTTL(rName, 2000),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, t, resourceName, &api2),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "OPENID_CONNECT"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.0.iat_ttl", "2000"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.0.issuer", "https://example.com"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAppSyncGraphQLAPI_OpenIDConnect_issuer(t *testing.T) {
	ctx := acctest.Context(t)
	var api1, api2 awstypes.GraphqlApi
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphQLAPIDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_openIDConnectIssuer(rName, "https://example.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, t, resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "OPENID_CONNECT"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.0.issuer", "https://example.com"),
				),
			},
			{
				Config: testAccGraphQLAPIConfig_openIDConnectIssuer(rName, "https://example.org"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, t, resourceName, &api2),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "OPENID_CONNECT"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.0.issuer", "https://example.org"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAppSyncGraphQLAPI_name(t *testing.T) {
	ctx := acctest.Context(t)
	var api1, api2 awstypes.GraphqlApi
	rName1 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName2 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphQLAPIDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_authenticationType(rName1, "API_KEY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, t, resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName1),
				),
			},
			{
				Config: testAccGraphQLAPIConfig_authenticationType(rName2, "API_KEY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, t, resourceName, &api2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
				),
			},
		},
	})
}

func TestAccAppSyncGraphQLAPI_UserPool_region(t *testing.T) {
	ctx := acctest.Context(t)
	var api1, api2 awstypes.GraphqlApi
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	cognitoUserPoolResourceName := "aws_cognito_user_pool.test"
	resourceName := "aws_appsync_graphql_api.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphQLAPIDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_userPoolRegion(rName, acctest.Region()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, t, resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "AMAZON_COGNITO_USER_POOLS"),
					resource.TestCheckResourceAttr(resourceName, "user_pool_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "user_pool_config.0.aws_region", acctest.Region()),
					resource.TestCheckResourceAttr(resourceName, "user_pool_config.0.default_action", "ALLOW"),
					resource.TestCheckResourceAttrPair(resourceName, "user_pool_config.0.user_pool_id", cognitoUserPoolResourceName, names.AttrID),
				),
			},
			{
				Config: testAccGraphQLAPIConfig_userPoolDefaultAction(rName, "ALLOW"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, t, resourceName, &api2),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "AMAZON_COGNITO_USER_POOLS"),
					resource.TestCheckResourceAttr(resourceName, "user_pool_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "user_pool_config.0.aws_region", acctest.Region()),
					resource.TestCheckResourceAttr(resourceName, "user_pool_config.0.default_action", "ALLOW"),
					resource.TestCheckResourceAttrPair(resourceName, "user_pool_config.0.user_pool_id", cognitoUserPoolResourceName, names.AttrID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAppSyncGraphQLAPI_UserPool_defaultAction(t *testing.T) {
	ctx := acctest.Context(t)
	var api1, api2 awstypes.GraphqlApi
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	cognitoUserPoolResourceName := "aws_cognito_user_pool.test"
	resourceName := "aws_appsync_graphql_api.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphQLAPIDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_userPoolDefaultAction(rName, "ALLOW"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, t, resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "AMAZON_COGNITO_USER_POOLS"),
					resource.TestCheckResourceAttr(resourceName, "user_pool_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "user_pool_config.0.aws_region", acctest.Region()),
					resource.TestCheckResourceAttr(resourceName, "user_pool_config.0.default_action", "ALLOW"),
					resource.TestCheckResourceAttrPair(resourceName, "user_pool_config.0.user_pool_id", cognitoUserPoolResourceName, names.AttrID),
				),
			},
			{
				Config: testAccGraphQLAPIConfig_userPoolDefaultAction(rName, "DENY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, t, resourceName, &api2),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "AMAZON_COGNITO_USER_POOLS"),
					resource.TestCheckResourceAttr(resourceName, "user_pool_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "user_pool_config.0.aws_region", acctest.Region()),
					resource.TestCheckResourceAttr(resourceName, "user_pool_config.0.default_action", "DENY"),
					resource.TestCheckResourceAttrPair(resourceName, "user_pool_config.0.user_pool_id", cognitoUserPoolResourceName, names.AttrID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAppSyncGraphQLAPI_LambdaAuthorizerConfig_authorizerURI(t *testing.T) {
	ctx := acctest.Context(t)
	var api1, api2 awstypes.GraphqlApi
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"
	lambdaAuthorizerResourceName := "aws_lambda_function.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphQLAPIDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_lambdaAuthorizerAuthorizerURI(rName, "aws_lambda_function.test.arn"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, t, resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "AWS_LAMBDA"),
					resource.TestCheckResourceAttr(resourceName, "lambda_authorizer_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_authorizer_config.0.authorizer_uri", lambdaAuthorizerResourceName, names.AttrARN),
				),
			},
			{
				Config: testAccGraphQLAPIConfig_lambdaAuthorizerAuthorizerURI(rName, "aws_lambda_function.test.qualified_arn"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, t, resourceName, &api2),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "AWS_LAMBDA"),
					resource.TestCheckResourceAttr(resourceName, "lambda_authorizer_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_authorizer_config.0.authorizer_uri", lambdaAuthorizerResourceName, "qualified_arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAppSyncGraphQLAPI_LambdaAuthorizerConfig_identityValidationExpression(t *testing.T) {
	ctx := acctest.Context(t)
	var api1, api2 awstypes.GraphqlApi
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"
	lambdaAuthorizerResourceName := "aws_lambda_function.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphQLAPIDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_lambdaAuthorizerIdentityValidationExpression(rName, "^test1$"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, t, resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "AWS_LAMBDA"),
					resource.TestCheckResourceAttr(resourceName, "lambda_authorizer_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_authorizer_config.0.authorizer_uri", lambdaAuthorizerResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "lambda_authorizer_config.0.identity_validation_expression", "^test1$"),
				),
			},
			{
				Config: testAccGraphQLAPIConfig_lambdaAuthorizerIdentityValidationExpression(rName, "^test2$"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, t, resourceName, &api2),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "AWS_LAMBDA"),
					resource.TestCheckResourceAttr(resourceName, "lambda_authorizer_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_authorizer_config.0.authorizer_uri", lambdaAuthorizerResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "lambda_authorizer_config.0.identity_validation_expression", "^test2$"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAppSyncGraphQLAPI_LambdaAuthorizerConfig_authorizerResultTTLInSeconds(t *testing.T) {
	ctx := acctest.Context(t)
	var api1, api2, api3, api4 awstypes.GraphqlApi
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphQLAPIDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_lambdaAuthorizerAuthorizerURI(rName, "aws_lambda_function.test.arn"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, t, resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "AWS_LAMBDA"),
					resource.TestCheckResourceAttr(resourceName, "lambda_authorizer_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "lambda_authorizer_config.0.authorizer_result_ttl_in_seconds", strconv.Itoa(tfappsync.DefaultAuthorizerResultTTLInSeconds)),
				),
			},
			{
				Config: testAccGraphQLAPIConfig_lambdaAuthorizerAuthorizerResultTTLInSeconds(rName, "123"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, t, resourceName, &api2),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "AWS_LAMBDA"),
					resource.TestCheckResourceAttr(resourceName, "lambda_authorizer_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "lambda_authorizer_config.0.authorizer_result_ttl_in_seconds", "123"),
				),
			},
			{
				Config: testAccGraphQLAPIConfig_lambdaAuthorizerAuthorizerResultTTLInSeconds(rName, "0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, t, resourceName, &api3),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "AWS_LAMBDA"),
					resource.TestCheckResourceAttr(resourceName, "lambda_authorizer_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "lambda_authorizer_config.0.authorizer_result_ttl_in_seconds", "0"),
				),
			},
			{
				Config: testAccGraphQLAPIConfig_lambdaAuthorizerAuthorizerURI(rName, "aws_lambda_function.test.arn"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, t, resourceName, &api4),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "AWS_LAMBDA"),
					resource.TestCheckResourceAttr(resourceName, "lambda_authorizer_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "lambda_authorizer_config.0.authorizer_result_ttl_in_seconds", strconv.Itoa(tfappsync.DefaultAuthorizerResultTTLInSeconds)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAppSyncGraphQLAPI_AdditionalAuthentication_apiKey(t *testing.T) {
	ctx := acctest.Context(t)
	var api1 awstypes.GraphqlApi
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphQLAPIDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_additionalAuthAuthType(rName, "AWS_IAM", "API_KEY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, t, resourceName, &api1),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "appsync", regexache.MustCompile(`apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "AWS_IAM"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.authentication_type", "API_KEY"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.openid_connect_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.user_pool_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.lambda_authorizer_config.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAppSyncGraphQLAPI_AdditionalAuthentication_iam(t *testing.T) {
	ctx := acctest.Context(t)
	var api1 awstypes.GraphqlApi
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphQLAPIDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_additionalAuthAuthType(rName, "API_KEY", "AWS_IAM"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, t, resourceName, &api1),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "appsync", regexache.MustCompile(`apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "API_KEY"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.authentication_type", "AWS_IAM"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.openid_connect_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.user_pool_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.lambda_authorizer_config.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAppSyncGraphQLAPI_AdditionalAuthentication_cognitoUserPools(t *testing.T) {
	ctx := acctest.Context(t)
	var api1 awstypes.GraphqlApi
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	cognitoUserPoolResourceName := "aws_cognito_user_pool.test"
	resourceName := "aws_appsync_graphql_api.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphQLAPIDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_additionalAuthUserPool(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, t, resourceName, &api1),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "appsync", regexache.MustCompile(`apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "API_KEY"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.authentication_type", "AMAZON_COGNITO_USER_POOLS"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.openid_connect_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.lambda_authorizer_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.user_pool_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "additional_authentication_provider.0.user_pool_config.0.user_pool_id", cognitoUserPoolResourceName, names.AttrID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAppSyncGraphQLAPI_AdditionalAuthentication_openIDConnect(t *testing.T) {
	ctx := acctest.Context(t)
	var api1 awstypes.GraphqlApi
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphQLAPIDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_additionalAuthOpenIdConnect(rName, "https://example.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, t, resourceName, &api1),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "appsync", regexache.MustCompile(`apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "API_KEY"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.authentication_type", "OPENID_CONNECT"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.user_pool_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.lambda_authorizer_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.openid_connect_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.openid_connect_config.0.issuer", "https://example.com"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAppSyncGraphQLAPI_AdditionalAuthentication_lambda(t *testing.T) {
	ctx := acctest.Context(t)
	var api1 awstypes.GraphqlApi
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"
	lambdaAuthorizerResourceName := "aws_lambda_function.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphQLAPIDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_additionalAuthLambda(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, t, resourceName, &api1),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "appsync", regexache.MustCompile(`apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "API_KEY"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.authentication_type", "AWS_LAMBDA"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.openid_connect_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.user_pool_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.lambda_authorizer_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "additional_authentication_provider.0.lambda_authorizer_config.0.authorizer_uri", lambdaAuthorizerResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.lambda_authorizer_config.0.authorizer_result_ttl_in_seconds", strconv.Itoa(tfappsync.DefaultAuthorizerResultTTLInSeconds)),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.lambda_authorizer_config.0.identity_validation_expression", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAppSyncGraphQLAPI_AdditionalAuthentication_multiple(t *testing.T) {
	ctx := acctest.Context(t)
	var api1 awstypes.GraphqlApi
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	cognitoUserPoolResourceName := "aws_cognito_user_pool.test"
	lambdaAuthorizerResourceName := "aws_lambda_function.test"
	resourceName := "aws_appsync_graphql_api.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphQLAPIDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_additionalAuthMultiple(rName, "https://example.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, t, resourceName, &api1),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "appsync", regexache.MustCompile(`apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "API_KEY"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.#", "4"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.authentication_type", "AWS_IAM"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.user_pool_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.openid_connect_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.lambda_authorizer_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.1.authentication_type", "AMAZON_COGNITO_USER_POOLS"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.1.openid_connect_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.1.lambda_authorizer_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.1.user_pool_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "additional_authentication_provider.1.user_pool_config.0.user_pool_id", cognitoUserPoolResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.2.authentication_type", "OPENID_CONNECT"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.2.user_pool_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.2.lambda_authorizer_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.2.openid_connect_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.2.openid_connect_config.0.issuer", "https://example.com"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.3.authentication_type", "AWS_LAMBDA"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.3.openid_connect_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.3.user_pool_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.3.lambda_authorizer_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "additional_authentication_provider.3.lambda_authorizer_config.0.authorizer_uri", lambdaAuthorizerResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAppSyncGraphQLAPI_xrayEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	var api1, api2 awstypes.GraphqlApi
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphQLAPIDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_xrayEnabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, t, resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "xray_enabled", acctest.CtTrue),
				),
			},
			{
				Config: testAccGraphQLAPIConfig_xrayEnabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, t, resourceName, &api2),
					resource.TestCheckResourceAttr(resourceName, "xray_enabled", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccAppSyncGraphQLAPI_visibility(t *testing.T) {
	ctx := acctest.Context(t)
	var api1 awstypes.GraphqlApi
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphQLAPIDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_visibility(rName, "PRIVATE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, t, resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "visibility", "PRIVATE"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAppSyncGraphQLAPI_introspectionConfig(t *testing.T) {
	ctx := acctest.Context(t)
	var api1 awstypes.GraphqlApi
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphQLAPIDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_introspectionConfig(rName, "DISABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, t, resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "introspection_config", "DISABLED"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAppSyncGraphQLAPI_queryDepthLimit(t *testing.T) {
	ctx := acctest.Context(t)
	var api1 awstypes.GraphqlApi
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphQLAPIDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_queryDepthLimit(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, t, resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "query_depth_limit", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAppSyncGraphQLAPI_resolverCountLimit(t *testing.T) {
	ctx := acctest.Context(t)
	var api1 awstypes.GraphqlApi
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphQLAPIDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_resolverCountLimit(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, t, resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "resolver_count_limit", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckGraphQLAPIDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).AppSyncClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_appsync_graphql_api" {
				continue
			}

			_, err := tfappsync.FindGraphQLAPIByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("AppSync GraphQL API %s still exists", rs.Primary.ID)
		}
		return nil
	}
}

func testAccCheckGraphQLAPIExists(ctx context.Context, t *testing.T, n string, v *awstypes.GraphqlApi) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).AppSyncClient(ctx)

		output, err := tfappsync.FindGraphQLAPIByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckGraphQLAPITypeExists(ctx context.Context, t *testing.T, n, typeName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).AppSyncClient(ctx)

		_, err := tfappsync.FindTypeByThreePartKey(ctx, conn, rs.Primary.ID, awstypes.TypeDefinitionFormatSdl, typeName)

		return err
	}
}

func TestAccAppSyncGraphQLAPI_apiType(t *testing.T) {
	ctx := acctest.Context(t)
	var api1 awstypes.GraphqlApi
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphQLAPIDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_apiType(rName, string(awstypes.GraphQLApiTypeMerged)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, t, resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "api_type", string(awstypes.GraphQLApiTypeMerged)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccGraphQLAPIConfig_authenticationType(rName, authenticationType string) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = %[1]q
  name                = %[2]q
}
`, authenticationType, rName)
}

func testAccGraphQLAPIConfig_visibility(rName, visibility string) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %[1]q
  visibility          = %[2]q
}
`, rName, visibility)
}

func testAccGraphQLAPIConfig_enhancedMetricsConfig(rName, dataSourceLevelMetricsBehavior, operationLevelMetricsConfig, resolverLevelMetricsBehavior string) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %[1]q
  enhanced_metrics_config {
    data_source_level_metrics_behavior = %[2]q
    operation_level_metrics_config     = %[3]q
    resolver_level_metrics_behavior    = %[4]q
  }
}
`, rName, dataSourceLevelMetricsBehavior, operationLevelMetricsConfig, resolverLevelMetricsBehavior)
}

func testAccGraphQLAPIConfig_logFieldLogLevel(rName, fieldLogLevel string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<POLICY
{
    "Version": "2012-10-17",
    "Statement": [
        {
        "Effect": "Allow",
        "Principal": {
            "Service": "appsync.amazonaws.com"
        },
        "Action": "sts:AssumeRole"
        }
    ]
}
POLICY
}

resource "aws_iam_role_policy_attachment" "test" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSAppSyncPushToCloudWatchLogs"
  role       = aws_iam_role.test.name
}

resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %[1]q

  log_config {
    cloudwatch_logs_role_arn = aws_iam_role.test.arn
    field_log_level          = %[2]q
  }
}
`, rName, fieldLogLevel)
}

func testAccGraphQLAPIConfig_logExcludeVerboseContent(rName string, excludeVerboseContent bool) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<POLICY
{
    "Version": "2012-10-17",
    "Statement": [
        {
        "Effect": "Allow",
        "Principal": {
            "Service": "appsync.amazonaws.com"
        },
        "Action": "sts:AssumeRole"
        }
    ]
}
POLICY
}

resource "aws_iam_role_policy_attachment" "test" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSAppSyncPushToCloudWatchLogs"
  role       = aws_iam_role.test.name
}

resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %[1]q

  log_config {
    cloudwatch_logs_role_arn = aws_iam_role.test.arn
    field_log_level          = "ALL"
    exclude_verbose_content  = %[2]t
  }
}
`, rName, excludeVerboseContent)
}

func testAccGraphQLAPIConfig_openIDConnectAuthTTL(rName string, authTTL int) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "OPENID_CONNECT"
  name                = %[1]q

  openid_connect_config {
    auth_ttl = %[2]d
    issuer   = "https://example.com"
  }
}
`, rName, authTTL)
}

func testAccGraphQLAPIConfig_openIDConnectClientID(rName, clientID string) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "OPENID_CONNECT"
  name                = %[1]q

  openid_connect_config {
    client_id = %[2]q
    issuer    = "https://example.com"
  }
}
`, rName, clientID)
}

func testAccGraphQLAPIConfig_openIDConnectIatTTL(rName string, iatTTL int) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "OPENID_CONNECT"
  name                = %[1]q

  openid_connect_config {
    iat_ttl = %[2]d
    issuer  = "https://example.com"
  }
}
`, rName, iatTTL)
}

func testAccGraphQLAPIConfig_openIDConnectIssuer(rName, issuer string) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "OPENID_CONNECT"
  name                = %[1]q

  openid_connect_config {
    issuer = %[2]q
  }
}
`, rName, issuer)
}

func testAccGraphQLAPIConfig_userPoolRegion(rName, awsRegion string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q
}

resource "aws_appsync_graphql_api" "test" {
  authentication_type = "AMAZON_COGNITO_USER_POOLS"
  name                = %[1]q

  user_pool_config {
    aws_region     = %[2]q
    default_action = "ALLOW"
    user_pool_id   = aws_cognito_user_pool.test.id
  }
}
`, rName, awsRegion)
}

func testAccGraphQLAPIConfig_userPoolDefaultAction(rName, defaultAction string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q
}

resource "aws_appsync_graphql_api" "test" {
  authentication_type = "AMAZON_COGNITO_USER_POOLS"
  name                = %[1]q

  user_pool_config {
    default_action = %[2]q
    user_pool_id   = aws_cognito_user_pool.test.id
  }
}
`, rName, defaultAction)
}

func testAccGraphQLAPIConfig_LambdaAuthorizerConfig_base(rName string) string {
	return fmt.Sprintf(`
data "aws_iam_policy_document" "lambda_assume_role_policy" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.lambda_assume_role_policy.json
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  handler       = "lambdatest.handler"
  role          = aws_iam_role.test.arn
  runtime       = "nodejs20.x"
  publish       = true
}

resource "aws_lambda_permission" "test" {
  statement_id  = "appsync_lambda_authorizer"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.test.arn
  principal     = "appsync.amazonaws.com"
  source_arn    = aws_appsync_graphql_api.test.arn
}
`, rName)
}

func testAccGraphQLAPIConfig_lambdaAuthorizerAuthorizerURI(rName, authorizerUri string) string {
	return acctest.ConfigCompose(testAccGraphQLAPIConfig_LambdaAuthorizerConfig_base(rName), fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "AWS_LAMBDA"
  name                = %[1]q

  lambda_authorizer_config {
    authorizer_uri = %[2]s
  }
}
`, rName, authorizerUri))
}

func testAccGraphQLAPIConfig_lambdaAuthorizerIdentityValidationExpression(rName, identityValidationExpression string) string {
	return acctest.ConfigCompose(testAccGraphQLAPIConfig_LambdaAuthorizerConfig_base(rName), fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "AWS_LAMBDA"
  name                = %[1]q

  lambda_authorizer_config {
    authorizer_uri                 = aws_lambda_function.test.arn
    identity_validation_expression = %[2]q
  }
}
`, rName, identityValidationExpression))
}

func testAccGraphQLAPIConfig_lambdaAuthorizerAuthorizerResultTTLInSeconds(rName, authorizerResultTtlInSeconds string) string {
	return acctest.ConfigCompose(testAccGraphQLAPIConfig_LambdaAuthorizerConfig_base(rName), fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "AWS_LAMBDA"
  name                = %[1]q

  lambda_authorizer_config {
    authorizer_uri                   = aws_lambda_function.test.arn
    authorizer_result_ttl_in_seconds = %[2]q
  }
}
`, rName, authorizerResultTtlInSeconds))
}

func testAccGraphQLAPIConfig_schema(rName string) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %[1]q
  schema              = "type Mutation {\n\tputPost(id: ID!, title: String!): Post\n}\n\ntype Post {\n\tid: ID!\n\ttitle: String!\n}\n\ntype Query {\n\tsinglePost(id: ID!): Post\n}\n\nschema {\n\tquery: Query\n\tmutation: Mutation\n\n}\n"
}
`, rName)
}

func testAccGraphQLAPIConfig_schemaUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %[1]q
  schema              = "type Mutation {\n\tputPostV2(id: ID!, title: String!): PostV2\n}\n\ntype PostV2 {\n\tid: ID!\n\ttitle: String!\n}\n\ntype Query {\n\tsinglePostV2(id: ID!): PostV2\n}\n\nschema {\n\tquery: Query\n\tmutation: Mutation\n\n}\n"
}
`, rName)
}

func testAccGraphQLAPIConfig_additionalAuthAuthType(rName, defaultAuthType, additionalAuthType string) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = %[1]q
  name                = %[2]q

  additional_authentication_provider {
    authentication_type = %[3]q
  }
}`, defaultAuthType, rName, additionalAuthType)
}

func testAccGraphQLAPIConfig_additionalAuthUserPool(rName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q
}

resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %[1]q

  additional_authentication_provider {
    authentication_type = "AMAZON_COGNITO_USER_POOLS"

    user_pool_config {
      user_pool_id = aws_cognito_user_pool.test.id
    }
  }
}
`, rName)
}

func testAccGraphQLAPIConfig_additionalAuthOpenIdConnect(rName, issuer string) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %[1]q

  additional_authentication_provider {
    authentication_type = "OPENID_CONNECT"

    openid_connect_config {
      issuer = %[2]q
    }
  }
}
`, rName, issuer)
}

func testAccGraphQLAPIConfig_additionalAuthLambda(rName string) string {
	return acctest.ConfigCompose(testAccGraphQLAPIConfig_LambdaAuthorizerConfig_base(rName), fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %[1]q

  additional_authentication_provider {
    authentication_type = "AWS_LAMBDA"

    lambda_authorizer_config {
      authorizer_uri = aws_lambda_function.test.arn
    }
  }
}
`, rName))
}

func testAccGraphQLAPIConfig_additionalAuthMultiple(rName, issuer string) string {
	return acctest.ConfigCompose(testAccGraphQLAPIConfig_LambdaAuthorizerConfig_base(rName), fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q
}

resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %[1]q

  additional_authentication_provider {
    authentication_type = "AWS_IAM"
  }

  additional_authentication_provider {
    authentication_type = "AMAZON_COGNITO_USER_POOLS"

    user_pool_config {
      user_pool_id = aws_cognito_user_pool.test.id
    }
  }

  additional_authentication_provider {
    authentication_type = "OPENID_CONNECT"

    openid_connect_config {
      issuer = %[2]q
    }
  }

  additional_authentication_provider {
    authentication_type = "AWS_LAMBDA"

    lambda_authorizer_config {
      authorizer_uri = aws_lambda_function.test.arn
    }
  }
}
`, rName, issuer))
}

func testAccGraphQLAPIConfig_xrayEnabled(rName string, xrayEnabled bool) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %[1]q
  xray_enabled        = %[2]t
}
`, rName, xrayEnabled)
}

func testAccGraphQLAPIConfig_introspectionConfig(rName, introspectionConfig string) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type  = "API_KEY"
  name                 = %[1]q
  introspection_config = %[2]q
}
`, rName, introspectionConfig)
}

func testAccGraphQLAPIConfig_queryDepthLimit(rName string, queryDepthLimit int) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %[1]q
  query_depth_limit   = %[2]d
}
`, rName, queryDepthLimit)
}

func testAccGraphQLAPIConfig_resolverCountLimit(rName string, resolverCountLimit int) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type  = "API_KEY"
  name                 = %[1]q
  resolver_count_limit = %[2]d
}
`, rName, resolverCountLimit)
}

func testAccGraphQLAPIConfig_apiType(rName, apiType string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  assume_role_policy = data.aws_iam_policy_document.test.json
  name_prefix        = %[1]q
}

data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

data "aws_region" "current" {}

data "aws_iam_policy_document" "test" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      identifiers = ["appsync.amazonaws.com"]
      type        = "Service"
    }
    condition {
      test     = "StringEquals"
      values   = [data.aws_caller_identity.current.account_id]
      variable = "aws:SourceAccount"
    }

    condition {
      test     = "ArnLike"
      values   = ["arn:${data.aws_partition.current.partition}:appsync:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}::apis/*"]
      variable = "aws:SourceArn"
    }
  }
}

resource "aws_appsync_graphql_api" "test" {
  authentication_type           = "API_KEY"
  name                          = %[1]q
  api_type                      = %[2]q
  merged_api_execution_role_arn = aws_iam_role.test.arn
}
`, rName, apiType)
}
