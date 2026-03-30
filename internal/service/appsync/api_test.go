// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package appsync_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/appsync/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfappsync "github.com/hashicorp/terraform-provider-aws/internal/service/appsync"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAppSyncAPI_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var api awstypes.Api
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_appsync_api.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAPIDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAPIConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAPIExists(ctx, t, resourceName, &api),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "event_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "event_config.0.auth_provider.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "event_config.0.auth_provider.0.auth_type", "API_KEY"),
					resource.TestCheckResourceAttr(resourceName, "event_config.0.auth_provider.0.cognito_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "event_config.0.auth_provider.0.lambda_authorizer_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "event_config.0.auth_provider.0.openid_connect_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "event_config.0.connection_auth_mode.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "event_config.0.connection_auth_mode.0.auth_type", "API_KEY"),
					resource.TestCheckResourceAttr(resourceName, "event_config.0.default_publish_auth_mode.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "event_config.0.default_publish_auth_mode.0.auth_type", "API_KEY"),
					resource.TestCheckResourceAttr(resourceName, "event_config.0.default_subscribe_auth_mode.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "event_config.0.default_subscribe_auth_mode.0.auth_type", "API_KEY"),
					resource.TestCheckResourceAttr(resourceName, "event_config.0.log_config.#", "0"),
					resource.TestCheckNoResourceAttr(resourceName, "owner_contact"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "api_id",
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "api_id"),
			},
		},
	})
}

func TestAccAppSyncAPI_comprehensive(t *testing.T) {
	ctx := acctest.Context(t)
	var api awstypes.Api
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_appsync_api.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAPIDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAPIConfig_eventConfigComprehensive(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAPIExists(ctx, t, resourceName, &api),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "owner_contact", acctest.DefaultEmailAddress),
					resource.TestCheckResourceAttr(resourceName, "event_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "event_config.0.auth_provider.#", "3"),
					// Cognito auth provider
					resource.TestCheckResourceAttr(resourceName, "event_config.0.auth_provider.0.auth_type", "AMAZON_COGNITO_USER_POOLS"),
					resource.TestCheckResourceAttr(resourceName, "event_config.0.auth_provider.0.cognito_config.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "event_config.0.auth_provider.0.cognito_config.0.user_pool_id"),
					resource.TestCheckResourceAttrSet(resourceName, "event_config.0.auth_provider.0.cognito_config.0.aws_region"),
					// Lambda auth provider
					resource.TestCheckResourceAttr(resourceName, "event_config.0.auth_provider.1.auth_type", "AWS_LAMBDA"),
					resource.TestCheckResourceAttr(resourceName, "event_config.0.auth_provider.1.lambda_authorizer_config.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "event_config.0.auth_provider.1.lambda_authorizer_config.0.authorizer_uri"),
					resource.TestCheckResourceAttr(resourceName, "event_config.0.auth_provider.1.lambda_authorizer_config.0.authorizer_result_ttl_in_seconds", "300"),
					// OpenID Connect auth provider
					resource.TestCheckResourceAttr(resourceName, "event_config.0.auth_provider.2.auth_type", "OPENID_CONNECT"),
					resource.TestCheckResourceAttr(resourceName, "event_config.0.auth_provider.2.openid_connect_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "event_config.0.auth_provider.2.openid_connect_config.0.issuer", "https://example.com"),
					resource.TestCheckResourceAttr(resourceName, "event_config.0.auth_provider.2.openid_connect_config.0.client_id", "test-client-id"),
					// Auth modes
					resource.TestCheckResourceAttr(resourceName, "event_config.0.connection_auth_mode.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "event_config.0.connection_auth_mode.0.auth_type", "AWS_LAMBDA"),
					resource.TestCheckResourceAttr(resourceName, "event_config.0.default_publish_auth_mode.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "event_config.0.default_publish_auth_mode.0.auth_type", "AWS_LAMBDA"),
					resource.TestCheckResourceAttr(resourceName, "event_config.0.default_subscribe_auth_mode.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "event_config.0.default_subscribe_auth_mode.0.auth_type", "AWS_LAMBDA"),
					// Log config
					resource.TestCheckResourceAttr(resourceName, "event_config.0.log_config.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "event_config.0.log_config.0.cloudwatch_logs_role_arn"),
					resource.TestCheckResourceAttr(resourceName, "event_config.0.log_config.0.log_level", "ERROR"),
					// Tags
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Environment", "test"),
					resource.TestCheckResourceAttr(resourceName, "tags.Purpose", "event-api-testing"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.Environment", "test"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.Purpose", "event-api-testing"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "api_id",
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "api_id"),
			},
		},
	})
}

func TestAccAppSyncAPI_update(t *testing.T) {
	ctx := acctest.Context(t)
	var api awstypes.Api
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rNameUpdated := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_appsync_api.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAPIDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAPIConfig_eventConfigComprehensive(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAPIExists(ctx, t, resourceName, &api),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "owner_contact", acctest.DefaultEmailAddress),
					resource.TestCheckResourceAttr(resourceName, "event_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "event_config.0.auth_provider.#", "3"),
				),
			},
			{
				Config: testAccAPIConfig_eventConfigComprehensive(rNameUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAPIExists(ctx, t, resourceName, &api),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, "owner_contact", acctest.DefaultEmailAddress),
					resource.TestCheckResourceAttr(resourceName, "event_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "event_config.0.auth_provider.#", "3"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "api_id",
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "api_id"),
			},
		},
	})
}
func TestAccAppSyncAPI_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var api awstypes.Api
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_appsync_api.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAPIDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAPIConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAPIExists(ctx, t, resourceName, &api),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfappsync.ResourceAPI, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func testAccCheckAPIDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).AppSyncClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_appsync_api" {
				continue
			}

			_, err := tfappsync.FindAPIByID(ctx, conn, rs.Primary.Attributes["api_id"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("AppSync API %s still exists", rs.Primary.Attributes["api_id"])
		}

		return nil
	}
}

func testAccCheckAPIExists(ctx context.Context, t *testing.T, n string, v *awstypes.Api) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).AppSyncClient(ctx)

		output, err := tfappsync.FindAPIByID(ctx, conn, rs.Primary.Attributes["api_id"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccAPIConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_appsync_api" "test" {
  name = %[1]q

  event_config {
    auth_provider {
      auth_type = "API_KEY"
    }

    connection_auth_mode {
      auth_type = "API_KEY"
    }

    default_publish_auth_mode {
      auth_type = "API_KEY"
    }

    default_subscribe_auth_mode {
      auth_type = "API_KEY"
    }
  }
}
`, rName)
}

func testAccAPIConfig_eventConfigComprehensive(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_cognito_user_pool" "test" {
  name = %[1]q
}

resource "aws_iam_role" "lambda" {
  name = "%[1]s-lambda"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      }
    ]
  })
}

data "aws_iam_policy_document" "lambda_basic" {
  statement {
    effect = "Allow"
    actions = [
      "logs:CreateLogGroup",
      "logs:CreateLogStream",
      "logs:PutLogEvents"
    ]
    resources = ["arn:${data.aws_partition.current.partition}:logs:*:*:*"]
  }
}

resource "aws_iam_role_policy" "lambda_basic" {
  name   = "%[1]s-lambda-basic"
  role   = aws_iam_role.lambda.id
  policy = data.aws_iam_policy_document.lambda_basic.json
}

resource "aws_lambda_function" "test" {
  filename         = "test-fixtures/lambdatest.zip"
  function_name    = %[1]q
  role             = aws_iam_role.lambda.arn
  handler          = "index.handler"
  runtime          = "nodejs18.x"
  source_code_hash = filebase64sha256("test-fixtures/lambdatest.zip")

  depends_on = [aws_iam_role_policy.lambda_basic]
}

resource "aws_lambda_permission" "appsync_invoke" {
  statement_id  = "AllowExecutionFromAppSync"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.test.function_name
  principal     = "appsync.amazonaws.com"
}

resource "aws_iam_role" "cloudwatch" {
  name = "%[1]s-cloudwatch"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "appsync.amazonaws.com"
        }
      }
    ]
  })
}

data "aws_iam_policy_document" "cloudwatch_logs" {
  statement {
    effect = "Allow"
    actions = [
      "logs:CreateLogGroup",
      "logs:CreateLogStream",
      "logs:PutLogEvents"
    ]
    resources = ["arn:${data.aws_partition.current.partition}:logs:*:*:*"]
  }
}

resource "aws_iam_role_policy" "cloudwatch" {
  name   = "%[1]s-cloudwatch-policy"
  role   = aws_iam_role.cloudwatch.id
  policy = data.aws_iam_policy_document.cloudwatch_logs.json
}

resource "aws_appsync_api" "test" {
  name          = %[1]q
  owner_contact = %[2]q

  event_config {
    auth_provider {
      auth_type = "AMAZON_COGNITO_USER_POOLS"
      cognito_config {
        user_pool_id = aws_cognito_user_pool.test.id
        aws_region   = data.aws_region.current.name
      }
    }

    auth_provider {
      auth_type = "AWS_LAMBDA"
      lambda_authorizer_config {
        authorizer_uri                   = aws_lambda_function.test.arn
        authorizer_result_ttl_in_seconds = 300
      }
    }

    auth_provider {
      auth_type = "OPENID_CONNECT"
      openid_connect_config {
        issuer    = "https://example.com"
        client_id = "test-client-id"
      }
    }

    connection_auth_mode {
      auth_type = "AWS_LAMBDA"
    }

    default_publish_auth_mode {
      auth_type = "AWS_LAMBDA"
    }

    default_subscribe_auth_mode {
      auth_type = "AWS_LAMBDA"
    }

    log_config {
      cloudwatch_logs_role_arn = aws_iam_role.cloudwatch.arn
      log_level                = "ERROR"
    }
  }

  tags = {
    Environment = "test"
    Purpose     = "event-api-testing"
  }

  depends_on = [
    aws_lambda_permission.appsync_invoke,
    aws_iam_role_policy.cloudwatch
  ]
}

data "aws_region" "current" {}
`, rName, acctest.DefaultEmailAddress)
}
