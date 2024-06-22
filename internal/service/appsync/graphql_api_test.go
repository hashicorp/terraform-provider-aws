// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appsync_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appsync/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfappsync "github.com/hashicorp/terraform-provider-aws/internal/service/appsync"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccGraphQLAPI_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var api1 awstypes.GraphqlApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphQLAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_authenticationType(rName, "API_KEY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, resourceName, &api1),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "appsync", regexache.MustCompile(`apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "API_KEY"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "log_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "user_pool_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "lambda_authorizer_config.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "uris.%"),
					resource.TestCheckResourceAttrSet(resourceName, "uris.GRAPHQL"),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrTags),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "xray_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "visibility", "GLOBAL"),
					resource.TestCheckResourceAttr(resourceName, "introspection_config", "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "query_depth_limit", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "resolver_count_limit", acctest.Ct0),
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

func testAccGraphQLAPI_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var api1 awstypes.GraphqlApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphQLAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_authenticationType(rName, "API_KEY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, resourceName, &api1),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfappsync.ResourceGraphQLAPI(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccGraphQLAPI_schema(t *testing.T) {
	ctx := acctest.Context(t)
	var api1, api2 awstypes.GraphqlApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphQLAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_schema(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, resourceName, &api1),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "appsync", regexache.MustCompile(`apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "API_KEY"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "log_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "user_pool_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "lambda_authorizer_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "xray_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrSchema),
					resource.TestCheckResourceAttrSet(resourceName, "uris.%"),
					resource.TestCheckResourceAttrSet(resourceName, "uris.GRAPHQL"),
					testAccCheckGraphQLAPITypeExists(ctx, resourceName, "Post"),
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
					testAccCheckGraphQLAPIExists(ctx, resourceName, &api2),
					testAccCheckGraphQLAPITypeExists(ctx, resourceName, "PostV2"),
				),
			},
		},
	})
}

func testAccGraphQLAPI_authenticationType(t *testing.T) {
	ctx := acctest.Context(t)
	var api1, api2 awstypes.GraphqlApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphQLAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_authenticationType(rName, "API_KEY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "API_KEY"),
				),
			},
			{
				Config: testAccGraphQLAPIConfig_authenticationType(rName, "AWS_IAM"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, resourceName, &api2),
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

func testAccGraphQLAPI_AuthenticationType_apiKey(t *testing.T) {
	ctx := acctest.Context(t)
	var api1 awstypes.GraphqlApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphQLAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_authenticationType(rName, "API_KEY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, resourceName, &api1),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "appsync", regexache.MustCompile(`apis/.+`)),
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

func testAccGraphQLAPI_AuthenticationType_iam(t *testing.T) {
	ctx := acctest.Context(t)
	var api1 awstypes.GraphqlApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphQLAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_authenticationType(rName, "AWS_IAM"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, resourceName, &api1),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "appsync", regexache.MustCompile(`apis/.+`)),
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

func testAccGraphQLAPI_AuthenticationType_amazonCognitoUserPools(t *testing.T) {
	ctx := acctest.Context(t)
	var api1 awstypes.GraphqlApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	cognitoUserPoolResourceName := "aws_cognito_user_pool.test"
	resourceName := "aws_appsync_graphql_api.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphQLAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_userPoolDefaultAction(rName, "ALLOW"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "AMAZON_COGNITO_USER_POOLS"),
					resource.TestCheckResourceAttr(resourceName, "user_pool_config.#", acctest.Ct1),
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

func testAccGraphQLAPI_AuthenticationType_openIDConnect(t *testing.T) {
	ctx := acctest.Context(t)
	var api1 awstypes.GraphqlApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphQLAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_openIDConnectIssuer(rName, "https://example.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "OPENID_CONNECT"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.#", acctest.Ct1),
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

func testAccGraphQLAPI_AuthenticationType_lambda(t *testing.T) {
	ctx := acctest.Context(t)
	var api1 awstypes.GraphqlApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"
	lambdaAuthorizerResourceName := "aws_lambda_function.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphQLAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_lambdaAuthorizerAuthorizerURI(rName, "aws_lambda_function.test.arn"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "AWS_LAMBDA"),
					resource.TestCheckResourceAttr(resourceName, "lambda_authorizer_config.#", acctest.Ct1),
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

func testAccGraphQLAPI_log(t *testing.T) {
	ctx := acctest.Context(t)
	var api1 awstypes.GraphqlApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_appsync_graphql_api.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphQLAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_logFieldLogLevel(rName, "ALL"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "log_config.#", acctest.Ct1),
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

func testAccGraphQLAPI_Log_fieldLogLevel(t *testing.T) {
	ctx := acctest.Context(t)
	var api1, api2, api3 awstypes.GraphqlApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_appsync_graphql_api.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphQLAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_logFieldLogLevel(rName, "ALL"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "log_config.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "log_config.0.cloudwatch_logs_role_arn", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "log_config.0.field_log_level", "ALL"),
					resource.TestCheckResourceAttr(resourceName, "log_config.0.exclude_verbose_content", acctest.CtFalse),
				),
			},
			{
				Config: testAccGraphQLAPIConfig_logFieldLogLevel(rName, "ERROR"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, resourceName, &api2),
					resource.TestCheckResourceAttr(resourceName, "log_config.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "log_config.0.cloudwatch_logs_role_arn", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "log_config.0.field_log_level", "ERROR"),
					resource.TestCheckResourceAttr(resourceName, "log_config.0.exclude_verbose_content", acctest.CtFalse),
				),
			},
			{
				Config: testAccGraphQLAPIConfig_logFieldLogLevel(rName, "NONE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, resourceName, &api3),
					resource.TestCheckResourceAttr(resourceName, "log_config.#", acctest.Ct1),
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

func testAccGraphQLAPI_Log_excludeVerboseContent(t *testing.T) {
	ctx := acctest.Context(t)
	var api1, api2 awstypes.GraphqlApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_appsync_graphql_api.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphQLAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_logExcludeVerboseContent(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "log_config.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "log_config.0.cloudwatch_logs_role_arn", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "log_config.0.field_log_level", "ALL"),
					resource.TestCheckResourceAttr(resourceName, "log_config.0.exclude_verbose_content", acctest.CtFalse),
				),
			},
			{
				Config: testAccGraphQLAPIConfig_logExcludeVerboseContent(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, resourceName, &api2),
					resource.TestCheckResourceAttr(resourceName, "log_config.#", acctest.Ct1),
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

func testAccGraphQLAPI_OpenIDConnect_authTTL(t *testing.T) {
	ctx := acctest.Context(t)
	var api1, api2 awstypes.GraphqlApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphQLAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_openIDConnectAuthTTL(rName, 1000),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "OPENID_CONNECT"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.0.auth_ttl", "1000"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.0.issuer", "https://example.com"),
				),
			},
			{
				Config: testAccGraphQLAPIConfig_openIDConnectAuthTTL(rName, 2000),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, resourceName, &api2),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "OPENID_CONNECT"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.#", acctest.Ct1),
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

func testAccGraphQLAPI_OpenIDConnect_clientID(t *testing.T) {
	ctx := acctest.Context(t)
	var api1, api2 awstypes.GraphqlApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphQLAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_openIDConnectClientID(rName, "ClientID1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "OPENID_CONNECT"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.0.client_id", "ClientID1"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.0.issuer", "https://example.com"),
				),
			},
			{
				Config: testAccGraphQLAPIConfig_openIDConnectClientID(rName, "ClientID2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, resourceName, &api2),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "OPENID_CONNECT"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.#", acctest.Ct1),
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

func testAccGraphQLAPI_OpenIDConnect_iatTTL(t *testing.T) {
	ctx := acctest.Context(t)
	var api1, api2 awstypes.GraphqlApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphQLAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_openIDConnectIatTTL(rName, 1000),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "OPENID_CONNECT"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.0.iat_ttl", "1000"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.0.issuer", "https://example.com"),
				),
			},
			{
				Config: testAccGraphQLAPIConfig_openIDConnectIatTTL(rName, 2000),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, resourceName, &api2),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "OPENID_CONNECT"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.#", acctest.Ct1),
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

func testAccGraphQLAPI_OpenIDConnect_issuer(t *testing.T) {
	ctx := acctest.Context(t)
	var api1, api2 awstypes.GraphqlApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphQLAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_openIDConnectIssuer(rName, "https://example.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "OPENID_CONNECT"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.0.issuer", "https://example.com"),
				),
			},
			{
				Config: testAccGraphQLAPIConfig_openIDConnectIssuer(rName, "https://example.org"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, resourceName, &api2),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "OPENID_CONNECT"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.#", acctest.Ct1),
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

func testAccGraphQLAPI_name(t *testing.T) {
	ctx := acctest.Context(t)
	var api1, api2 awstypes.GraphqlApi
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphQLAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_authenticationType(rName1, "API_KEY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName1),
				),
			},
			{
				Config: testAccGraphQLAPIConfig_authenticationType(rName2, "API_KEY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, resourceName, &api2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
				),
			},
		},
	})
}

func testAccGraphQLAPI_UserPool_region(t *testing.T) {
	ctx := acctest.Context(t)
	var api1, api2 awstypes.GraphqlApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	cognitoUserPoolResourceName := "aws_cognito_user_pool.test"
	resourceName := "aws_appsync_graphql_api.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphQLAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_userPoolRegion(rName, acctest.Region()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "AMAZON_COGNITO_USER_POOLS"),
					resource.TestCheckResourceAttr(resourceName, "user_pool_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "user_pool_config.0.aws_region", acctest.Region()),
					resource.TestCheckResourceAttr(resourceName, "user_pool_config.0.default_action", "ALLOW"),
					resource.TestCheckResourceAttrPair(resourceName, "user_pool_config.0.user_pool_id", cognitoUserPoolResourceName, names.AttrID),
				),
			},
			{
				Config: testAccGraphQLAPIConfig_userPoolDefaultAction(rName, "ALLOW"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, resourceName, &api2),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "AMAZON_COGNITO_USER_POOLS"),
					resource.TestCheckResourceAttr(resourceName, "user_pool_config.#", acctest.Ct1),
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

func testAccGraphQLAPI_UserPool_defaultAction(t *testing.T) {
	ctx := acctest.Context(t)
	var api1, api2 awstypes.GraphqlApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	cognitoUserPoolResourceName := "aws_cognito_user_pool.test"
	resourceName := "aws_appsync_graphql_api.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphQLAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_userPoolDefaultAction(rName, "ALLOW"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "AMAZON_COGNITO_USER_POOLS"),
					resource.TestCheckResourceAttr(resourceName, "user_pool_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "user_pool_config.0.aws_region", acctest.Region()),
					resource.TestCheckResourceAttr(resourceName, "user_pool_config.0.default_action", "ALLOW"),
					resource.TestCheckResourceAttrPair(resourceName, "user_pool_config.0.user_pool_id", cognitoUserPoolResourceName, names.AttrID),
				),
			},
			{
				Config: testAccGraphQLAPIConfig_userPoolDefaultAction(rName, "DENY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, resourceName, &api2),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "AMAZON_COGNITO_USER_POOLS"),
					resource.TestCheckResourceAttr(resourceName, "user_pool_config.#", acctest.Ct1),
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

func testAccGraphQLAPI_LambdaAuthorizerConfig_authorizerURI(t *testing.T) {
	ctx := acctest.Context(t)
	var api1, api2 awstypes.GraphqlApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"
	lambdaAuthorizerResourceName := "aws_lambda_function.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphQLAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_lambdaAuthorizerAuthorizerURI(rName, "aws_lambda_function.test.arn"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "AWS_LAMBDA"),
					resource.TestCheckResourceAttr(resourceName, "lambda_authorizer_config.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_authorizer_config.0.authorizer_uri", lambdaAuthorizerResourceName, names.AttrARN),
				),
			},
			{
				Config: testAccGraphQLAPIConfig_lambdaAuthorizerAuthorizerURI(rName, "aws_lambda_function.test.qualified_arn"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, resourceName, &api2),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "AWS_LAMBDA"),
					resource.TestCheckResourceAttr(resourceName, "lambda_authorizer_config.#", acctest.Ct1),
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

func testAccGraphQLAPI_LambdaAuthorizerConfig_identityValidationExpression(t *testing.T) {
	ctx := acctest.Context(t)
	var api1, api2 awstypes.GraphqlApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"
	lambdaAuthorizerResourceName := "aws_lambda_function.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphQLAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_lambdaAuthorizerIdentityValidationExpression(rName, "^test1$"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "AWS_LAMBDA"),
					resource.TestCheckResourceAttr(resourceName, "lambda_authorizer_config.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_authorizer_config.0.authorizer_uri", lambdaAuthorizerResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "lambda_authorizer_config.0.identity_validation_expression", "^test1$"),
				),
			},
			{
				Config: testAccGraphQLAPIConfig_lambdaAuthorizerIdentityValidationExpression(rName, "^test2$"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, resourceName, &api2),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "AWS_LAMBDA"),
					resource.TestCheckResourceAttr(resourceName, "lambda_authorizer_config.#", acctest.Ct1),
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

func testAccGraphQLAPI_LambdaAuthorizerConfig_authorizerResultTTLInSeconds(t *testing.T) {
	ctx := acctest.Context(t)
	var api1, api2, api3, api4 awstypes.GraphqlApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphQLAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_lambdaAuthorizerAuthorizerURI(rName, "aws_lambda_function.test.arn"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "AWS_LAMBDA"),
					resource.TestCheckResourceAttr(resourceName, "lambda_authorizer_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "lambda_authorizer_config.0.authorizer_result_ttl_in_seconds", strconv.Itoa(tfappsync.DefaultAuthorizerResultTTLInSeconds)),
				),
			},
			{
				Config: testAccGraphQLAPIConfig_lambdaAuthorizerAuthorizerResultTTLInSeconds(rName, "123"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, resourceName, &api2),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "AWS_LAMBDA"),
					resource.TestCheckResourceAttr(resourceName, "lambda_authorizer_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "lambda_authorizer_config.0.authorizer_result_ttl_in_seconds", "123"),
				),
			},
			{
				Config: testAccGraphQLAPIConfig_lambdaAuthorizerAuthorizerResultTTLInSeconds(rName, acctest.Ct0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, resourceName, &api3),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "AWS_LAMBDA"),
					resource.TestCheckResourceAttr(resourceName, "lambda_authorizer_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "lambda_authorizer_config.0.authorizer_result_ttl_in_seconds", acctest.Ct0),
				),
			},
			{
				Config: testAccGraphQLAPIConfig_lambdaAuthorizerAuthorizerURI(rName, "aws_lambda_function.test.arn"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, resourceName, &api4),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "AWS_LAMBDA"),
					resource.TestCheckResourceAttr(resourceName, "lambda_authorizer_config.#", acctest.Ct1),
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

func testAccGraphQLAPI_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var api1 awstypes.GraphqlApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphQLAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGraphQLAPIConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccGraphQLAPIConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccGraphQLAPI_AdditionalAuthentication_apiKey(t *testing.T) {
	ctx := acctest.Context(t)
	var api1 awstypes.GraphqlApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphQLAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_additionalAuthAuthType(rName, "AWS_IAM", "API_KEY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, resourceName, &api1),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "appsync", regexache.MustCompile(`apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "AWS_IAM"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.authentication_type", "API_KEY"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.openid_connect_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.user_pool_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.lambda_authorizer_config.#", acctest.Ct0),
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

func testAccGraphQLAPI_AdditionalAuthentication_iam(t *testing.T) {
	ctx := acctest.Context(t)
	var api1 awstypes.GraphqlApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphQLAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_additionalAuthAuthType(rName, "API_KEY", "AWS_IAM"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, resourceName, &api1),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "appsync", regexache.MustCompile(`apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "API_KEY"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.authentication_type", "AWS_IAM"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.openid_connect_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.user_pool_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.lambda_authorizer_config.#", acctest.Ct0),
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

func testAccGraphQLAPI_AdditionalAuthentication_cognitoUserPools(t *testing.T) {
	ctx := acctest.Context(t)
	var api1 awstypes.GraphqlApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	cognitoUserPoolResourceName := "aws_cognito_user_pool.test"
	resourceName := "aws_appsync_graphql_api.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphQLAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_additionalAuthUserPool(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, resourceName, &api1),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "appsync", regexache.MustCompile(`apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "API_KEY"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.authentication_type", "AMAZON_COGNITO_USER_POOLS"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.openid_connect_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.lambda_authorizer_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.user_pool_config.#", acctest.Ct1),
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

func testAccGraphQLAPI_AdditionalAuthentication_openIDConnect(t *testing.T) {
	ctx := acctest.Context(t)
	var api1 awstypes.GraphqlApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphQLAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_additionalAuthOpenIdConnect(rName, "https://example.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, resourceName, &api1),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "appsync", regexache.MustCompile(`apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "API_KEY"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.authentication_type", "OPENID_CONNECT"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.user_pool_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.lambda_authorizer_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.openid_connect_config.#", acctest.Ct1),
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

func testAccGraphQLAPI_AdditionalAuthentication_lambda(t *testing.T) {
	ctx := acctest.Context(t)
	var api1 awstypes.GraphqlApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"
	lambdaAuthorizerResourceName := "aws_lambda_function.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphQLAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_additionalAuthLambda(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, resourceName, &api1),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "appsync", regexache.MustCompile(`apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "API_KEY"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.authentication_type", "AWS_LAMBDA"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.openid_connect_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.user_pool_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.lambda_authorizer_config.#", acctest.Ct1),
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

func testAccGraphQLAPI_AdditionalAuthentication_multiple(t *testing.T) {
	ctx := acctest.Context(t)
	var api1 awstypes.GraphqlApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	cognitoUserPoolResourceName := "aws_cognito_user_pool.test"
	lambdaAuthorizerResourceName := "aws_lambda_function.test"
	resourceName := "aws_appsync_graphql_api.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphQLAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_additionalAuthMultiple(rName, "https://example.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, resourceName, &api1),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "appsync", regexache.MustCompile(`apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "API_KEY"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.#", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.authentication_type", "AWS_IAM"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.user_pool_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.openid_connect_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.lambda_authorizer_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.1.authentication_type", "AMAZON_COGNITO_USER_POOLS"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.1.openid_connect_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.1.lambda_authorizer_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.1.user_pool_config.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "additional_authentication_provider.1.user_pool_config.0.user_pool_id", cognitoUserPoolResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.2.authentication_type", "OPENID_CONNECT"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.2.user_pool_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.2.lambda_authorizer_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.2.openid_connect_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.2.openid_connect_config.0.issuer", "https://example.com"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.3.authentication_type", "AWS_LAMBDA"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.3.openid_connect_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.3.user_pool_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.3.lambda_authorizer_config.#", acctest.Ct1),
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

func testAccGraphQLAPI_xrayEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	var api1, api2 awstypes.GraphqlApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphQLAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_xrayEnabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "xray_enabled", acctest.CtTrue),
				),
			},
			{
				Config: testAccGraphQLAPIConfig_xrayEnabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, resourceName, &api2),
					resource.TestCheckResourceAttr(resourceName, "xray_enabled", acctest.CtFalse),
				),
			},
		},
	})
}

func testAccGraphQLAPI_visibility(t *testing.T) {
	ctx := acctest.Context(t)
	var api1 awstypes.GraphqlApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphQLAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_visibility(rName, "PRIVATE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, resourceName, &api1),
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

func testAccGraphQLAPI_introspectionConfig(t *testing.T) {
	ctx := acctest.Context(t)
	var api1 awstypes.GraphqlApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphQLAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_introspectionConfig(rName, "DISABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, resourceName, &api1),
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

func testAccGraphQLAPI_queryDepthLimit(t *testing.T) {
	ctx := acctest.Context(t)
	var api1 awstypes.GraphqlApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphQLAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_queryDepthLimit(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "query_depth_limit", acctest.Ct2),
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

func testAccGraphQLAPI_resolverCountLimit(t *testing.T) {
	ctx := acctest.Context(t)
	var api1 awstypes.GraphqlApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGraphQLAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_resolverCountLimit(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(ctx, resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "resolver_count_limit", acctest.Ct2),
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

func testAccCheckGraphQLAPIDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AppSyncClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_appsync_graphql_api" {
				continue
			}

			_, err := tfappsync.FindGraphQLAPIByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
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

func testAccCheckGraphQLAPIExists(ctx context.Context, n string, v *awstypes.GraphqlApi) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppSyncClient(ctx)

		output, err := tfappsync.FindGraphQLAPIByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckGraphQLAPITypeExists(ctx context.Context, n, typeName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppSyncClient(ctx)

		_, err := tfappsync.FindTypeByThreePartKey(ctx, conn, rs.Primary.ID, awstypes.TypeDefinitionFormatSdl, typeName)

		return err
	}
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

func testAccGraphQLAPIConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccGraphQLAPIConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
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
