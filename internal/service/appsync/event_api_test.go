// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appsync_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appsync/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"

	tfappsync "github.com/hashicorp/terraform-provider-aws/internal/service/appsync"
)

// TIP: File Structure. The basic outline for all test files should be as
// follows. Improve this resource's maintainability by following this
// outline.
//
// 1. Package declaration (add "_test" since this is a test file)
// 2. Imports
// 3. Unit tests
// 4. Basic test
// 5. Disappears test
// 6. All the other tests
// 7. Helper functions (exists, destroy, check, etc.)
// 8. Functions that return Terraform configurations

// TIP: ==== UNIT TESTS ====
// This is an example of a unit test. Its name is not prefixed with
// "TestAcc" like an acceptance test.
//
// Unlike acceptance tests, unit tests do not access AWS and are focused on a
// function (or method). Because of this, they are quick and cheap to run.
//
// In designing a resource's implementation, isolate complex bits from AWS bits
// so that they can be tested through a unit test. We encourage more unit tests
// in the provider.
//
// Cut and dry functions using well-used patterns, like typical flatteners and
// expanders, don't need unit testing. However, if they are complex or
// intricate, they should be unit tested.
// Unit tests would go here if needed

// TIP: ==== ACCEPTANCE TESTS ====
// This is an example of a basic acceptance test. This should test as much of
// standard functionality of the resource as possible, and test importing, if
// applicable. We prefix its name with "TestAcc", the service, and the
// resource name.
//
// Acceptance test access AWS and cost money to run.
func TestAccAppSyncEventApi_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var eventapi awstypes.Api
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appsync_event_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventApiDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventApiConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEventApiExists(ctx, resourceName, &eventapi),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, "api_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "appsync", regexache.MustCompile(`api/.+$`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
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

func TestAccAppSyncEventApi_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var eventapi awstypes.Api
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appsync_event_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventApiDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventApiConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEventApiExists(ctx, resourceName, &eventapi),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfappsync.ResourceEventApi, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAppSyncEventApi_eventConfigComprehensive(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var eventapi awstypes.Api
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appsync_event_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventApiDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventApiConfig_eventConfigComprehensive(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEventApiExists(ctx, resourceName, &eventapi),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "owner_contact", "test@example.com"),
					resource.TestCheckResourceAttr(resourceName, "event_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "event_config.0.auth_providers.#", "3"),
					// Cognito auth provider
					resource.TestCheckResourceAttr(resourceName, "event_config.0.auth_providers.0.auth_type", "AMAZON_COGNITO_USER_POOLS"),
					resource.TestCheckResourceAttr(resourceName, "event_config.0.auth_providers.0.cognito_config.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "event_config.0.auth_providers.0.cognito_config.0.user_pool_id"),
					resource.TestCheckResourceAttrSet(resourceName, "event_config.0.auth_providers.0.cognito_config.0.aws_region"),
					// Lambda auth provider
					resource.TestCheckResourceAttr(resourceName, "event_config.0.auth_providers.1.auth_type", "AWS_LAMBDA"),
					resource.TestCheckResourceAttr(resourceName, "event_config.0.auth_providers.1.lambda_authorizer_config.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "event_config.0.auth_providers.1.lambda_authorizer_config.0.authorizer_uri"),
					resource.TestCheckResourceAttr(resourceName, "event_config.0.auth_providers.1.lambda_authorizer_config.0.authorizer_result_ttl_in_seconds", "300"),
					// OpenID Connect auth provider
					resource.TestCheckResourceAttr(resourceName, "event_config.0.auth_providers.2.auth_type", "OPENID_CONNECT"),
					resource.TestCheckResourceAttr(resourceName, "event_config.0.auth_providers.2.openid_connect_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "event_config.0.auth_providers.2.openid_connect_config.0.issuer", "https://example.com"),
					resource.TestCheckResourceAttr(resourceName, "event_config.0.auth_providers.2.openid_connect_config.0.client_id", "test-client-id"),
					// Auth modes
					resource.TestCheckResourceAttr(resourceName, "event_config.0.connection_auth_modes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "event_config.0.connection_auth_modes.0.auth_type", "AWS_LAMBDA"),
					resource.TestCheckResourceAttr(resourceName, "event_config.0.default_publish_auth_modes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "event_config.0.default_publish_auth_modes.0.auth_type", "AWS_LAMBDA"),
					resource.TestCheckResourceAttr(resourceName, "event_config.0.default_subscribe_auth_modes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "event_config.0.default_subscribe_auth_modes.0.auth_type", "AWS_LAMBDA"),
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
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAppSyncEventApi_tags(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var eventapi awstypes.Api
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appsync_event_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventApiDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventApiConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEventApiExists(ctx, resourceName, &eventapi),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEventApiConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEventApiExists(ctx, resourceName, &eventapi),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key2", "value2"),
				),
			},
			{
				Config: testAccEventApiConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEventApiExists(ctx, resourceName, &eventapi),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAppSyncEventApi_update(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var eventapi awstypes.Api
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameUpdated := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appsync_event_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventApiDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventApiConfig_eventConfigComprehensive(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEventApiExists(ctx, resourceName, &eventapi),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "owner_contact", "test@example.com"),
					resource.TestCheckResourceAttr(resourceName, "event_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "event_config.0.auth_providers.#", "3"),
				),
			},
			{
				Config: testAccEventApiConfig_eventConfigComprehensive(rNameUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEventApiExists(ctx, resourceName, &eventapi),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, "owner_contact", "test@example.com"),
					resource.TestCheckResourceAttr(resourceName, "event_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "event_config.0.auth_providers.#", "3"),
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

func testAccCheckEventApiDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AppSyncClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_appsync_event_api" {
				continue
			}

			_, err := tfappsync.FindEventApiByID(ctx, conn, rs.Primary.ID)
			if tfresource.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.AppSync, create.ErrActionCheckingDestroyed, tfappsync.ResNameEventApi, rs.Primary.ID, err)
			}

			return create.Error(names.AppSync, create.ErrActionCheckingDestroyed, tfappsync.ResNameEventApi, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckEventApiExists(ctx context.Context, name string, eventapi *awstypes.Api) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.AppSync, create.ErrActionCheckingExistence, tfappsync.ResNameEventApi, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.AppSync, create.ErrActionCheckingExistence, tfappsync.ResNameEventApi, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppSyncClient(ctx)

		resp, err := tfappsync.FindEventApiByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.AppSync, create.ErrActionCheckingExistence, tfappsync.ResNameEventApi, rs.Primary.ID, err)
		}

		*eventapi = *resp

		return nil
	}
}

func testAccEventApiConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_appsync_event_api" "test" {
  name = %[1]q
}
`, rName)
}

func testAccEventApiConfig_eventConfigComprehensive(rName string) string {
	return fmt.Sprintf(`
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
    resources = ["arn:aws:logs:*:*:*"]
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
  role            = aws_iam_role.lambda.arn
  handler         = "index.handler"
  runtime         = "nodejs18.x"
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
    resources = ["arn:aws:logs:*:*:*"]
  }
}

resource "aws_iam_role_policy" "cloudwatch" {
  name   = "%[1]s-cloudwatch-policy"
  role   = aws_iam_role.cloudwatch.id
  policy = data.aws_iam_policy_document.cloudwatch_logs.json
}

resource "aws_appsync_event_api" "test" {
  name          = %[1]q
  owner_contact = "test@example.com"

  event_config {
    auth_providers {
      auth_type = "AMAZON_COGNITO_USER_POOLS"
      cognito_config {
        user_pool_id = aws_cognito_user_pool.test.id
        aws_region   = data.aws_region.current.name
      }
    }

    auth_providers {
      auth_type = "AWS_LAMBDA"
      lambda_authorizer_config {
        authorizer_uri                  = aws_lambda_function.test.arn
        authorizer_result_ttl_in_seconds = 300
      }
    }

    auth_providers {
      auth_type = "OPENID_CONNECT"
      openid_connect_config {
        issuer    = "https://example.com"
        client_id = "test-client-id"
      }
    }

    connection_auth_modes {
      auth_type = "AWS_LAMBDA"
    }

    default_publish_auth_modes {
      auth_type = "AWS_LAMBDA"
    }

    default_subscribe_auth_modes {
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
`, rName)
}

func testAccEventApiConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_appsync_event_api" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccEventApiConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_appsync_event_api" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
