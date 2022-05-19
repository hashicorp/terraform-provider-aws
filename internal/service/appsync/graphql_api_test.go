package appsync_test

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appsync"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfappsync "github.com/hashicorp/terraform-provider-aws/internal/service/appsync"
)

func testAccGraphQLAPI_basic(t *testing.T) {
	var api1 appsync.GraphqlApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appsync.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appsync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGraphQLAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_authenticationType(rName, "API_KEY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(resourceName, &api1),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "appsync", regexp.MustCompile(`apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "API_KEY"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "log_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "user_pool_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "lambda_authorizer_config.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "uris.%"),
					resource.TestCheckResourceAttrSet(resourceName, "uris.GRAPHQL"),
					resource.TestCheckNoResourceAttr(resourceName, "tags"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "xray_enabled", "false"),
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
	var api1 appsync.GraphqlApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appsync.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appsync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGraphQLAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_authenticationType(rName, "API_KEY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(resourceName, &api1),
					acctest.CheckResourceDisappears(acctest.Provider, tfappsync.ResourceGraphQLAPI(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccGraphQLAPI_schema(t *testing.T) {
	var api1, api2 appsync.GraphqlApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appsync.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appsync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGraphQLAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_schema(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(resourceName, &api1),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "appsync", regexp.MustCompile(`apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "API_KEY"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "log_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "user_pool_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "lambda_authorizer_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "xray_enabled", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "schema"),
					resource.TestCheckResourceAttrSet(resourceName, "uris.%"),
					resource.TestCheckResourceAttrSet(resourceName, "uris.GRAPHQL"),
					testAccCheckTypeExists(resourceName, "Post"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"schema"},
			},
			{
				Config: testAccGraphQLAPIConfig_schemaUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(resourceName, &api2),
					testAccCheckTypeExists(resourceName, "PostV2"),
				),
			},
		},
	})
}

func testAccGraphQLAPI_authenticationType(t *testing.T) {
	var api1, api2 appsync.GraphqlApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appsync.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appsync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGraphQLAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_authenticationType(rName, "API_KEY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "API_KEY"),
				),
			},
			{
				Config: testAccGraphQLAPIConfig_authenticationType(rName, "AWS_IAM"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(resourceName, &api2),
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
	var api1 appsync.GraphqlApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appsync.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appsync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGraphQLAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_authenticationType(rName, "API_KEY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(resourceName, &api1),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "appsync", regexp.MustCompile(`apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "API_KEY"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
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
	var api1 appsync.GraphqlApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appsync.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appsync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGraphQLAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_authenticationType(rName, "AWS_IAM"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(resourceName, &api1),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "appsync", regexp.MustCompile(`apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "AWS_IAM"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
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
	var api1 appsync.GraphqlApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	cognitoUserPoolResourceName := "aws_cognito_user_pool.test"
	resourceName := "aws_appsync_graphql_api.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appsync.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appsync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGraphQLAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_UserPoolConfig_defaultAction(rName, "ALLOW"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "AMAZON_COGNITO_USER_POOLS"),
					resource.TestCheckResourceAttr(resourceName, "user_pool_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "user_pool_config.0.aws_region", acctest.Region()),
					resource.TestCheckResourceAttr(resourceName, "user_pool_config.0.default_action", "ALLOW"),
					resource.TestCheckResourceAttrPair(resourceName, "user_pool_config.0.user_pool_id", cognitoUserPoolResourceName, "id"),
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
	var api1 appsync.GraphqlApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appsync.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appsync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGraphQLAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_OpenIDConnectConfig_issuer(rName, "https://example.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(resourceName, &api1),
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

func testAccGraphQLAPI_AuthenticationType_lambda(t *testing.T) {
	var api1 appsync.GraphqlApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"
	lambdaAuthorizerResourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appsync.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appsync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGraphQLAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_LambdaAuthorizerConfig_authorizerURI(rName, "aws_lambda_function.test.arn"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "AWS_LAMBDA"),
					resource.TestCheckResourceAttr(resourceName, "lambda_authorizer_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_authorizer_config.0.authorizer_uri", lambdaAuthorizerResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "lambda_authorizer_config.0.authorizer_result_ttl_in_seconds", strconv.Itoa(tfappsync.DefaultAuthorizerResultTtlInSeconds)),
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
	var api1 appsync.GraphqlApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_appsync_graphql_api.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appsync.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appsync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGraphQLAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_LogConfig_fieldLogLevel(rName, "ALL"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "log_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "log_config.0.cloudwatch_logs_role_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "log_config.0.field_log_level", "ALL"),
					resource.TestCheckResourceAttr(resourceName, "log_config.0.exclude_verbose_content", "false"),
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
	var api1, api2, api3 appsync.GraphqlApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_appsync_graphql_api.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appsync.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appsync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGraphQLAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_LogConfig_fieldLogLevel(rName, "ALL"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "log_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "log_config.0.cloudwatch_logs_role_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "log_config.0.field_log_level", "ALL"),
					resource.TestCheckResourceAttr(resourceName, "log_config.0.exclude_verbose_content", "false"),
				),
			},
			{
				Config: testAccGraphQLAPIConfig_LogConfig_fieldLogLevel(rName, "ERROR"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(resourceName, &api2),
					resource.TestCheckResourceAttr(resourceName, "log_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "log_config.0.cloudwatch_logs_role_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "log_config.0.field_log_level", "ERROR"),
					resource.TestCheckResourceAttr(resourceName, "log_config.0.exclude_verbose_content", "false"),
				),
			},
			{
				Config: testAccGraphQLAPIConfig_LogConfig_fieldLogLevel(rName, "NONE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(resourceName, &api3),
					resource.TestCheckResourceAttr(resourceName, "log_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "log_config.0.cloudwatch_logs_role_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "log_config.0.field_log_level", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "log_config.0.exclude_verbose_content", "false"),
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
	var api1, api2 appsync.GraphqlApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_appsync_graphql_api.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appsync.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appsync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGraphQLAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_LogConfig_excludeVerboseContent(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "log_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "log_config.0.cloudwatch_logs_role_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "log_config.0.field_log_level", "ALL"),
					resource.TestCheckResourceAttr(resourceName, "log_config.0.exclude_verbose_content", "false"),
				),
			},
			{
				Config: testAccGraphQLAPIConfig_LogConfig_excludeVerboseContent(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(resourceName, &api2),
					resource.TestCheckResourceAttr(resourceName, "log_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "log_config.0.cloudwatch_logs_role_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "log_config.0.field_log_level", "ALL"),
					resource.TestCheckResourceAttr(resourceName, "log_config.0.exclude_verbose_content", "true"),
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
	var api1, api2 appsync.GraphqlApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appsync.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appsync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGraphQLAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_OpenIDConnectConfig_authTTL(rName, 1000),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "OPENID_CONNECT"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.0.auth_ttl", "1000"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.0.issuer", "https://example.com"),
				),
			},
			{
				Config: testAccGraphQLAPIConfig_OpenIDConnectConfig_authTTL(rName, 2000),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(resourceName, &api2),
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

func testAccGraphQLAPI_OpenIDConnect_clientID(t *testing.T) {
	var api1, api2 appsync.GraphqlApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appsync.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appsync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGraphQLAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_OpenIDConnectConfig_clientID(rName, "ClientID1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "OPENID_CONNECT"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.0.client_id", "ClientID1"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.0.issuer", "https://example.com"),
				),
			},
			{
				Config: testAccGraphQLAPIConfig_OpenIDConnectConfig_clientID(rName, "ClientID2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(resourceName, &api2),
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

func testAccGraphQLAPI_OpenIDConnect_iatTTL(t *testing.T) {
	var api1, api2 appsync.GraphqlApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appsync.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appsync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGraphQLAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_OpenIDConnectConfig_iatTTL(rName, 1000),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "OPENID_CONNECT"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.0.iat_ttl", "1000"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.0.issuer", "https://example.com"),
				),
			},
			{
				Config: testAccGraphQLAPIConfig_OpenIDConnectConfig_iatTTL(rName, 2000),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(resourceName, &api2),
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

func testAccGraphQLAPI_OpenIDConnect_issuer(t *testing.T) {
	var api1, api2 appsync.GraphqlApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appsync.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appsync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGraphQLAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_OpenIDConnectConfig_issuer(rName, "https://example.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "OPENID_CONNECT"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.0.issuer", "https://example.com"),
				),
			},
			{
				Config: testAccGraphQLAPIConfig_OpenIDConnectConfig_issuer(rName, "https://example.org"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(resourceName, &api2),
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

func testAccGraphQLAPI_name(t *testing.T) {
	var api1, api2 appsync.GraphqlApi
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appsync.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appsync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGraphQLAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_authenticationType(rName1, "API_KEY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "name", rName1),
				),
			},
			{
				Config: testAccGraphQLAPIConfig_authenticationType(rName2, "API_KEY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(resourceName, &api2),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
				),
			},
		},
	})
}

func testAccGraphQLAPI_UserPool_region(t *testing.T) {
	var api1, api2 appsync.GraphqlApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	cognitoUserPoolResourceName := "aws_cognito_user_pool.test"
	resourceName := "aws_appsync_graphql_api.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appsync.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appsync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGraphQLAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_UserPoolConfig_region(rName, acctest.Region()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "AMAZON_COGNITO_USER_POOLS"),
					resource.TestCheckResourceAttr(resourceName, "user_pool_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "user_pool_config.0.aws_region", acctest.Region()),
					resource.TestCheckResourceAttr(resourceName, "user_pool_config.0.default_action", "ALLOW"),
					resource.TestCheckResourceAttrPair(resourceName, "user_pool_config.0.user_pool_id", cognitoUserPoolResourceName, "id"),
				),
			},
			{
				Config: testAccGraphQLAPIConfig_UserPoolConfig_defaultAction(rName, "ALLOW"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(resourceName, &api2),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "AMAZON_COGNITO_USER_POOLS"),
					resource.TestCheckResourceAttr(resourceName, "user_pool_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "user_pool_config.0.aws_region", acctest.Region()),
					resource.TestCheckResourceAttr(resourceName, "user_pool_config.0.default_action", "ALLOW"),
					resource.TestCheckResourceAttrPair(resourceName, "user_pool_config.0.user_pool_id", cognitoUserPoolResourceName, "id"),
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
	var api1, api2 appsync.GraphqlApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	cognitoUserPoolResourceName := "aws_cognito_user_pool.test"
	resourceName := "aws_appsync_graphql_api.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appsync.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appsync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGraphQLAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_UserPoolConfig_defaultAction(rName, "ALLOW"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "AMAZON_COGNITO_USER_POOLS"),
					resource.TestCheckResourceAttr(resourceName, "user_pool_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "user_pool_config.0.aws_region", acctest.Region()),
					resource.TestCheckResourceAttr(resourceName, "user_pool_config.0.default_action", "ALLOW"),
					resource.TestCheckResourceAttrPair(resourceName, "user_pool_config.0.user_pool_id", cognitoUserPoolResourceName, "id"),
				),
			},
			{
				Config: testAccGraphQLAPIConfig_UserPoolConfig_defaultAction(rName, "DENY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(resourceName, &api2),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "AMAZON_COGNITO_USER_POOLS"),
					resource.TestCheckResourceAttr(resourceName, "user_pool_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "user_pool_config.0.aws_region", acctest.Region()),
					resource.TestCheckResourceAttr(resourceName, "user_pool_config.0.default_action", "DENY"),
					resource.TestCheckResourceAttrPair(resourceName, "user_pool_config.0.user_pool_id", cognitoUserPoolResourceName, "id"),
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
	var api1, api2 appsync.GraphqlApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"
	lambdaAuthorizerResourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appsync.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appsync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGraphQLAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_LambdaAuthorizerConfig_authorizerURI(rName, "aws_lambda_function.test.arn"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "AWS_LAMBDA"),
					resource.TestCheckResourceAttr(resourceName, "lambda_authorizer_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_authorizer_config.0.authorizer_uri", lambdaAuthorizerResourceName, "arn"),
				),
			},
			{
				Config: testAccGraphQLAPIConfig_LambdaAuthorizerConfig_authorizerURI(rName, "aws_lambda_function.test.qualified_arn"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(resourceName, &api2),
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

func testAccGraphQLAPI_LambdaAuthorizerConfig_identityValidationExpression(t *testing.T) {
	var api1, api2 appsync.GraphqlApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"
	lambdaAuthorizerResourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appsync.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appsync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGraphQLAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_LambdaAuthorizerConfig_identityValidationExpression(rName, "^test1$"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "AWS_LAMBDA"),
					resource.TestCheckResourceAttr(resourceName, "lambda_authorizer_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_authorizer_config.0.authorizer_uri", lambdaAuthorizerResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "lambda_authorizer_config.0.identity_validation_expression", "^test1$"),
				),
			},
			{
				Config: testAccGraphQLAPIConfig_LambdaAuthorizerConfig_identityValidationExpression(rName, "^test2$"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(resourceName, &api2),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "AWS_LAMBDA"),
					resource.TestCheckResourceAttr(resourceName, "lambda_authorizer_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_authorizer_config.0.authorizer_uri", lambdaAuthorizerResourceName, "arn"),
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
	var api1, api2, api3, api4 appsync.GraphqlApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appsync.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appsync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGraphQLAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_LambdaAuthorizerConfig_authorizerURI(rName, "aws_lambda_function.test.arn"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "AWS_LAMBDA"),
					resource.TestCheckResourceAttr(resourceName, "lambda_authorizer_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "lambda_authorizer_config.0.authorizer_result_ttl_in_seconds", strconv.Itoa(tfappsync.DefaultAuthorizerResultTtlInSeconds)),
				),
			},
			{
				Config: testAccGraphQLAPIConfig_LambdaAuthorizerConfig_authorizerResultTTLInSeconds(rName, "123"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(resourceName, &api2),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "AWS_LAMBDA"),
					resource.TestCheckResourceAttr(resourceName, "lambda_authorizer_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "lambda_authorizer_config.0.authorizer_result_ttl_in_seconds", "123"),
				),
			},
			{
				Config: testAccGraphQLAPIConfig_LambdaAuthorizerConfig_authorizerResultTTLInSeconds(rName, "0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(resourceName, &api3),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "AWS_LAMBDA"),
					resource.TestCheckResourceAttr(resourceName, "lambda_authorizer_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "lambda_authorizer_config.0.authorizer_result_ttl_in_seconds", "0"),
				),
			},
			{
				Config: testAccGraphQLAPIConfig_LambdaAuthorizerConfig_authorizerURI(rName, "aws_lambda_function.test.arn"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(resourceName, &api4),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "AWS_LAMBDA"),
					resource.TestCheckResourceAttr(resourceName, "lambda_authorizer_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "lambda_authorizer_config.0.authorizer_result_ttl_in_seconds", strconv.Itoa(tfappsync.DefaultAuthorizerResultTtlInSeconds)),
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
	var api1 appsync.GraphqlApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appsync.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appsync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGraphQLAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_tags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "API_KEY"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value One"),
					resource.TestCheckResourceAttr(resourceName, "tags.Description", "Very interesting"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGraphQLAPIConfig_tagsModified(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value One Changed"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value Two"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "Value Three"),
				),
			},
		},
	})
}

func testAccGraphQLAPI_AdditionalAuthentication_apiKey(t *testing.T) {
	var api1 appsync.GraphqlApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appsync.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appsync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGraphQLAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_AdditionalAuth_authType(rName, "AWS_IAM", "API_KEY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(resourceName, &api1),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "appsync", regexp.MustCompile(`apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
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

func testAccGraphQLAPI_AdditionalAuthentication_iam(t *testing.T) {
	var api1 appsync.GraphqlApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appsync.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appsync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGraphQLAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_AdditionalAuth_authType(rName, "API_KEY", "AWS_IAM"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(resourceName, &api1),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "appsync", regexp.MustCompile(`apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
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

func testAccGraphQLAPI_AdditionalAuthentication_cognitoUserPools(t *testing.T) {
	var api1 appsync.GraphqlApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	cognitoUserPoolResourceName := "aws_cognito_user_pool.test"
	resourceName := "aws_appsync_graphql_api.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appsync.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appsync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGraphQLAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_AdditionalAuth_userPoolConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(resourceName, &api1),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "appsync", regexp.MustCompile(`apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "API_KEY"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.authentication_type", "AMAZON_COGNITO_USER_POOLS"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.openid_connect_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.lambda_authorizer_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.user_pool_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "additional_authentication_provider.0.user_pool_config.0.user_pool_id", cognitoUserPoolResourceName, "id"),
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
	var api1 appsync.GraphqlApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appsync.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appsync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGraphQLAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_AdditionalAuth_openIdConnect(rName, "https://example.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(resourceName, &api1),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "appsync", regexp.MustCompile(`apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
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
func testAccGraphQLAPI_AdditionalAuthentication_lambda(t *testing.T) {
	var api1 appsync.GraphqlApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"
	lambdaAuthorizerResourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appsync.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appsync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGraphQLAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_AdditionalAuth_lambda(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(resourceName, &api1),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "appsync", regexp.MustCompile(`apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "API_KEY"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.authentication_type", "AWS_LAMBDA"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.openid_connect_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.user_pool_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.lambda_authorizer_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "additional_authentication_provider.0.lambda_authorizer_config.0.authorizer_uri", lambdaAuthorizerResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.lambda_authorizer_config.0.authorizer_result_ttl_in_seconds", strconv.Itoa(tfappsync.DefaultAuthorizerResultTtlInSeconds)),
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
	var api1 appsync.GraphqlApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	cognitoUserPoolResourceName := "aws_cognito_user_pool.test"
	lambdaAuthorizerResourceName := "aws_lambda_function.test"
	resourceName := "aws_appsync_graphql_api.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appsync.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appsync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGraphQLAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_AdditionalAuth_multiple(rName, "https://example.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(resourceName, &api1),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "appsync", regexp.MustCompile(`apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
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
					resource.TestCheckResourceAttrPair(resourceName, "additional_authentication_provider.1.user_pool_config.0.user_pool_id", cognitoUserPoolResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.2.authentication_type", "OPENID_CONNECT"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.2.user_pool_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.2.lambda_authorizer_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.2.openid_connect_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.2.openid_connect_config.0.issuer", "https://example.com"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.3.authentication_type", "AWS_LAMBDA"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.3.openid_connect_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.3.user_pool_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.3.lambda_authorizer_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "additional_authentication_provider.3.lambda_authorizer_config.0.authorizer_uri", lambdaAuthorizerResourceName, "arn"),
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
	var api1, api2 appsync.GraphqlApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appsync_graphql_api.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appsync.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appsync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGraphQLAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLAPIConfig_xrayEnabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "xray_enabled", "true"),
				),
			},
			{
				Config: testAccGraphQLAPIConfig_xrayEnabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphQLAPIExists(resourceName, &api2),
					resource.TestCheckResourceAttr(resourceName, "xray_enabled", "false"),
				),
			},
		},
	})
}

func testAccCheckGraphQLAPIDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).AppSyncConn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_appsync_graphql_api" {
			continue
		}

		input := &appsync.GetGraphqlApiInput{
			ApiId: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetGraphqlApi(input)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, appsync.ErrCodeNotFoundException) {
				return nil
			}
			return err
		}
	}
	return nil
}

func testAccCheckGraphQLAPIExists(name string, api *appsync.GraphqlApi) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppSyncConn

		input := &appsync.GetGraphqlApiInput{
			ApiId: aws.String(rs.Primary.ID),
		}

		output, err := conn.GetGraphqlApi(input)

		if err != nil {
			return err
		}

		*api = *output.GraphqlApi

		return nil
	}
}

func testAccCheckTypeExists(name, typeName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppSyncConn

		input := &appsync.GetTypeInput{
			ApiId:    aws.String(rs.Primary.ID),
			TypeName: aws.String(typeName),
			Format:   aws.String(appsync.OutputTypeSdl),
		}

		_, err := conn.GetType(input)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccGraphQLAPIConfig_authenticationType(rName, authenticationType string) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = %q
  name                = %q
}
`, authenticationType, rName)
}

func testAccGraphQLAPIConfig_LogConfig_fieldLogLevel(rName, fieldLogLevel string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %q

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
  name                = %q

  log_config {
    cloudwatch_logs_role_arn = aws_iam_role.test.arn
    field_log_level          = %q
  }
}
`, rName, rName, fieldLogLevel)
}

func testAccGraphQLAPIConfig_LogConfig_excludeVerboseContent(rName string, excludeVerboseContent bool) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %q

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
  name                = %q

  log_config {
    cloudwatch_logs_role_arn = aws_iam_role.test.arn
    field_log_level          = "ALL"
    exclude_verbose_content  = %t
  }
}
`, rName, rName, excludeVerboseContent)
}

func testAccGraphQLAPIConfig_OpenIDConnectConfig_authTTL(rName string, authTTL int) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "OPENID_CONNECT"
  name                = %q

  openid_connect_config {
    auth_ttl = %d
    issuer   = "https://example.com"
  }
}
`, rName, authTTL)
}

func testAccGraphQLAPIConfig_OpenIDConnectConfig_clientID(rName, clientID string) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "OPENID_CONNECT"
  name                = %q

  openid_connect_config {
    client_id = %q
    issuer    = "https://example.com"
  }
}
`, rName, clientID)
}

func testAccGraphQLAPIConfig_OpenIDConnectConfig_iatTTL(rName string, iatTTL int) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "OPENID_CONNECT"
  name                = %q

  openid_connect_config {
    iat_ttl = %d
    issuer  = "https://example.com"
  }
}
`, rName, iatTTL)
}

func testAccGraphQLAPIConfig_OpenIDConnectConfig_issuer(rName, issuer string) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "OPENID_CONNECT"
  name                = %q

  openid_connect_config {
    issuer = %q
  }
}
`, rName, issuer)
}

func testAccGraphQLAPIConfig_UserPoolConfig_region(rName, awsRegion string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %q
}

resource "aws_appsync_graphql_api" "test" {
  authentication_type = "AMAZON_COGNITO_USER_POOLS"
  name                = %q

  user_pool_config {
    aws_region     = %q
    default_action = "ALLOW"
    user_pool_id   = aws_cognito_user_pool.test.id
  }
}
`, rName, rName, awsRegion)
}

func testAccGraphQLAPIConfig_UserPoolConfig_defaultAction(rName, defaultAction string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %q
}

resource "aws_appsync_graphql_api" "test" {
  authentication_type = "AMAZON_COGNITO_USER_POOLS"
  name                = %q

  user_pool_config {
    default_action = %q
    user_pool_id   = aws_cognito_user_pool.test.id
  }
}
`, rName, rName, defaultAction)
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
  runtime       = "nodejs14.x"
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

func testAccGraphQLAPIConfig_LambdaAuthorizerConfig_authorizerURI(rName, authorizerUri string) string {
	return acctest.ConfigCompose(testAccGraphQLAPIConfig_LambdaAuthorizerConfig_base(rName), fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "AWS_LAMBDA"
  name                = %q

  lambda_authorizer_config {
    authorizer_uri = %s
  }
}
`, rName, authorizerUri))
}

func testAccGraphQLAPIConfig_LambdaAuthorizerConfig_identityValidationExpression(rName, identityValidationExpression string) string {
	return acctest.ConfigCompose(testAccGraphQLAPIConfig_LambdaAuthorizerConfig_base(rName), fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "AWS_LAMBDA"
  name                = %q

  lambda_authorizer_config {
    authorizer_uri                 = aws_lambda_function.test.arn
    identity_validation_expression = %q
  }
}
`, rName, identityValidationExpression))
}

func testAccGraphQLAPIConfig_LambdaAuthorizerConfig_authorizerResultTTLInSeconds(rName, authorizerResultTtlInSeconds string) string {
	return acctest.ConfigCompose(testAccGraphQLAPIConfig_LambdaAuthorizerConfig_base(rName), fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "AWS_LAMBDA"
  name                = %q

  lambda_authorizer_config {
    authorizer_uri                   = aws_lambda_function.test.arn
    authorizer_result_ttl_in_seconds = %q
  }
}
`, rName, authorizerResultTtlInSeconds))
}

func testAccGraphQLAPIConfig_schema(rName string) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %q
  schema              = "type Mutation {\n\tputPost(id: ID!, title: String!): Post\n}\n\ntype Post {\n\tid: ID!\n\ttitle: String!\n}\n\ntype Query {\n\tsinglePost(id: ID!): Post\n}\n\nschema {\n\tquery: Query\n\tmutation: Mutation\n\n}\n"
}
`, rName)
}

func testAccGraphQLAPIConfig_schemaUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %q
  schema              = "type Mutation {\n\tputPostV2(id: ID!, title: String!): PostV2\n}\n\ntype PostV2 {\n\tid: ID!\n\ttitle: String!\n}\n\ntype Query {\n\tsinglePostV2(id: ID!): PostV2\n}\n\nschema {\n\tquery: Query\n\tmutation: Mutation\n\n}\n"
}
`, rName)
}

func testAccGraphQLAPIConfig_tags(rName string) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %q

  tags = {
    Key1        = "Value One"
    Description = "Very interesting"
  }
}
`, rName)
}

func testAccGraphQLAPIConfig_tagsModified(rName string) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %q

  tags = {
    Key1 = "Value One Changed"
    Key2 = "Value Two"
    Key3 = "Value Three"
  }
}
`, rName)
}

func testAccGraphQLAPIConfig_AdditionalAuth_authType(rName, defaultAuthType, additionalAuthType string) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = %q
  name                = %q

  additional_authentication_provider {
    authentication_type = %q
  }
}`, defaultAuthType, rName, additionalAuthType)
}

func testAccGraphQLAPIConfig_AdditionalAuth_userPoolConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %q
}

resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %q

  additional_authentication_provider {
    authentication_type = "AMAZON_COGNITO_USER_POOLS"

    user_pool_config {
      user_pool_id = aws_cognito_user_pool.test.id
    }
  }
}
`, rName, rName)
}

func testAccGraphQLAPIConfig_AdditionalAuth_openIdConnect(rName, issuer string) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %q

  additional_authentication_provider {
    authentication_type = "OPENID_CONNECT"

    openid_connect_config {
      issuer = %q
    }
  }
}
`, rName, issuer)
}

func testAccGraphQLAPIConfig_AdditionalAuth_lambda(rName string) string {
	return acctest.ConfigCompose(testAccGraphQLAPIConfig_LambdaAuthorizerConfig_base(rName), fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %q

  additional_authentication_provider {
    authentication_type = "AWS_LAMBDA"

    lambda_authorizer_config {
      authorizer_uri = aws_lambda_function.test.arn
    }
  }
}
`, rName))
}

func testAccGraphQLAPIConfig_AdditionalAuth_multiple(rName, issuer string) string {
	return acctest.ConfigCompose(testAccGraphQLAPIConfig_LambdaAuthorizerConfig_base(rName), fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %q
}

resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %q

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
      issuer = %q
    }
  }

  additional_authentication_provider {
    authentication_type = "AWS_LAMBDA"

    lambda_authorizer_config {
      authorizer_uri = aws_lambda_function.test.arn
    }
  }
}
`, rName, rName, issuer))
}

func testAccGraphQLAPIConfig_xrayEnabled(rName string, xrayEnabled bool) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %q
  xray_enabled        = %t
}
`, rName, xrayEnabled)
}
