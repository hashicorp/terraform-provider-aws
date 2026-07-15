// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package logs_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tflogs "github.com/hashicorp/terraform-provider-aws/internal/service/logs"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLogsSubscriptionFilter_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var filter types.SubscriptionFilter
	lambdaFunctionResourceName := "aws_lambda_function.test"
	logGroupResourceName := "aws_cloudwatch_log_group.test"
	resourceName := "aws_cloudwatch_log_subscription_filter.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubscriptionFilterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSubscriptionFilterConfig_destinationARNLambda(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubscriptionFilterExists(ctx, t, resourceName, &filter),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrDestinationARN, lambdaFunctionResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "distribution", "ByLogStream"),
					resource.TestCheckNoResourceAttr(resourceName, "emit_system_fields"),
					resource.TestCheckResourceAttr(resourceName, "filter_pattern", "logtype test"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrLogGroupName, logGroupResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateIdFunc:       testAccSubscriptionFilterImportStateIDFunc(resourceName),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrRoleARN, "apply_on_transformed_logs"},
			},
		},
	})
}

func TestAccLogsSubscriptionFilter_many(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_cloudwatch_log_subscription_filter.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubscriptionFilterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSubscriptionFilterConfig_destinationARNLambdaMany(rName, 2), // This is the default limit of subscription filters on an account
				Check:  testAccCheckSubscriptionFilterManyExists(ctx, t, resourceName, 2),
			},
		},
	})
}

func TestAccLogsSubscriptionFilter_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var filter types.SubscriptionFilter
	resourceName := "aws_cloudwatch_log_subscription_filter.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubscriptionFilterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSubscriptionFilterConfig_destinationARNLambda(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubscriptionFilterExists(ctx, t, resourceName, &filter),
					acctest.CheckSDKResourceDisappears(ctx, t, tflogs.ResourceSubscriptionFilter(), resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccLogsSubscriptionFilter_Disappears_logGroup(t *testing.T) {
	ctx := acctest.Context(t)
	var filter types.SubscriptionFilter
	logGroupResourceName := "aws_cloudwatch_log_group.test"
	resourceName := "aws_cloudwatch_log_subscription_filter.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubscriptionFilterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSubscriptionFilterConfig_destinationARNLambda(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubscriptionFilterExists(ctx, t, resourceName, &filter),
					acctest.CheckSDKResourceDisappears(ctx, t, tflogs.ResourceGroup(), logGroupResourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccLogsSubscriptionFilter_DestinationARN_kinesisDataFirehose(t *testing.T) {
	ctx := acctest.Context(t)
	var filter types.SubscriptionFilter
	firehoseResourceName := "aws_kinesis_firehose_delivery_stream.test"
	resourceName := "aws_cloudwatch_log_subscription_filter.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubscriptionFilterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSubscriptionFilterConfig_destinationARNKinesisDataFirehose(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubscriptionFilterExists(ctx, t, resourceName, &filter),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrDestinationARN, firehoseResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateIdFunc:       testAccSubscriptionFilterImportStateIDFunc(resourceName),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_on_transformed_logs"},
			},
		},
	})
}

func TestAccLogsSubscriptionFilter_DestinationARN_kinesisStream(t *testing.T) {
	ctx := acctest.Context(t)
	var filter types.SubscriptionFilter
	kinesisStream := "aws_kinesis_stream.test"
	resourceName := "aws_cloudwatch_log_subscription_filter.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubscriptionFilterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSubscriptionFilterConfig_destinationARNKinesisStream(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubscriptionFilterExists(ctx, t, resourceName, &filter),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrDestinationARN, kinesisStream, names.AttrARN),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateIdFunc:       testAccSubscriptionFilterImportStateIDFunc(resourceName),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_on_transformed_logs"},
			},
		},
	})
}

func TestAccLogsSubscriptionFilter_distribution(t *testing.T) {
	ctx := acctest.Context(t)
	var filter types.SubscriptionFilter
	resourceName := "aws_cloudwatch_log_subscription_filter.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubscriptionFilterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSubscriptionFilterConfig_distribution(rName, "Random"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubscriptionFilterExists(ctx, t, resourceName, &filter),
					resource.TestCheckResourceAttr(resourceName, "distribution", "Random"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateIdFunc:       testAccSubscriptionFilterImportStateIDFunc(resourceName),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrRoleARN, "apply_on_transformed_logs"},
			},
			{
				Config: testAccSubscriptionFilterConfig_distribution(rName, "ByLogStream"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubscriptionFilterExists(ctx, t, resourceName, &filter),
					resource.TestCheckResourceAttr(resourceName, "distribution", "ByLogStream"),
				),
			},
		},
	})
}

func TestAccLogsSubscriptionFilter_emitSystemFields(t *testing.T) {
	ctx := acctest.Context(t)
	var filter types.SubscriptionFilter
	resourceName := "aws_cloudwatch_log_subscription_filter.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubscriptionFilterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSubscriptionFilterConfig_emitSystemFields(rName, "[\"@aws.region\"]"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubscriptionFilterExists(ctx, t, resourceName, &filter),
					resource.TestCheckResourceAttr(resourceName, "emit_system_fields.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "emit_system_fields.0", "@aws.region"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateIdFunc:       testAccSubscriptionFilterImportStateIDFunc(resourceName),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrRoleARN, "apply_on_transformed_logs"},
			},
			{
				Config: testAccSubscriptionFilterConfig_emitSystemFields(rName, "[\"@aws.account\", \"@aws.region\"]"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubscriptionFilterExists(ctx, t, resourceName, &filter),
					resource.TestCheckResourceAttr(resourceName, "emit_system_fields.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "emit_system_fields.0", "@aws.account"),
					resource.TestCheckResourceAttr(resourceName, "emit_system_fields.1", "@aws.region"),
				),
			},
			{
				Config: testAccSubscriptionFilterConfig_emitSystemFields(rName, "[]"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubscriptionFilterExists(ctx, t, resourceName, &filter),
					resource.TestCheckResourceAttr(resourceName, "emit_system_fields.#", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateIdFunc:       testAccSubscriptionFilterImportStateIDFunc(resourceName),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrRoleARN, "apply_on_transformed_logs"},
			},
		},
	})
}

func TestAccLogsSubscriptionFilter_roleARN(t *testing.T) {
	ctx := acctest.Context(t)
	var filter types.SubscriptionFilter
	iamRoleResourceName1 := "aws_iam_role.test"
	iamRoleResourceName2 := "aws_iam_role.test2"
	resourceName := "aws_cloudwatch_log_subscription_filter.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubscriptionFilterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSubscriptionFilterConfig_roleARN1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubscriptionFilterExists(ctx, t, resourceName, &filter),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, iamRoleResourceName1, names.AttrARN),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateIdFunc:       testAccSubscriptionFilterImportStateIDFunc(resourceName),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_on_transformed_logs"},
			},
			{
				Config: testAccSubscriptionFilterConfig_roleARN2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubscriptionFilterExists(ctx, t, resourceName, &filter),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, iamRoleResourceName2, names.AttrARN),
				),
			},
		},
	})
}

func TestAccLogsSubscriptionFilter_applyOnTransformedLogs(t *testing.T) {
	ctx := acctest.Context(t)
	var filter types.SubscriptionFilter
	resourceName := "aws_cloudwatch_log_subscription_filter.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubscriptionFilterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				Config: testAccSubscriptionFilterConfig_applyOnTransformed(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubscriptionFilterExists(ctx, t, resourceName, &filter),
					resource.TestCheckResourceAttr(resourceName, "apply_on_transformed_logs", acctest.CtTrue),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateIdFunc:       testAccSubscriptionFilterImportStateIDFunc(resourceName),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrRoleARN},
			},
			{
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Config: testAccSubscriptionFilterConfig_applyOnTransformed(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubscriptionFilterExists(ctx, t, resourceName, &filter),
					resource.TestCheckResourceAttr(resourceName, "apply_on_transformed_logs", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccLogsSubscriptionFilter_DestinationARN_kinesisDataFirehose_crossAccount(t *testing.T) {
	ctx := acctest.Context(t)
	var filter types.SubscriptionFilter
	providers := make(map[string]*schema.Provider)

	clwLogDestinationResourceName := "aws_cloudwatch_log_destination.firehose"
	resourceName := "aws_cloudwatch_log_subscription_filter.firehose"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesNamedAlternate(ctx, t, providers),
		CheckDestroy:             testAccCheckSubscriptionFilterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				// Initialize the providers.
				Config: testAccSubscriptionFilterConfig_kinesisDataFirehoseCrossAccountBase_initProviders(),
			},
			{
				PreConfig: func() {
					// Can only run check here because the provider is not available until the previous step.
					acctest.PreCheckSameOrganization(ctx, t, acctest.DefaultProviderFunc, acctest.NamedProviderFunc(acctest.ProviderNameAlternate, providers))
				},
				Config: testAccSubscriptionFilterConfig_kinesisDataFirehoseCrossAccountBase(rName),
			},
			{
				// Apply the configuration in two separate steps to reproduce the IAM eventual consistency issue regarding policy application.
				// The subscription filter, the IAM role used by the filter and its IAM role policy are created in a dedicated step.
				// The policy will be attached to the role, giving CloudWatch Logs no time to observe the updated permissions before calling PutSubscriptionFilter.
				Config: testAccSubscriptionFilterConfig_destinationARNKinesisDataFirehoseCrossAccount(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubscriptionFilterExists(ctx, t, resourceName, &filter),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrDestinationARN, clwLogDestinationResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateIdFunc:       testAccSubscriptionFilterImportStateIDFunc(resourceName),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_on_transformed_logs"},
			},
		},
	})
}

func testAccCheckSubscriptionFilterDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).LogsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudwatch_log_subscription_filter" {
				continue
			}

			_, err := tflogs.FindSubscriptionFilterByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrLogGroupName], rs.Primary.Attributes[names.AttrName])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudWatch Logs Filter Subscription still exists: %s", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckSubscriptionFilterExists(ctx context.Context, t *testing.T, n string, v *types.SubscriptionFilter) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).LogsClient(ctx)

		output, err := tflogs.FindSubscriptionFilterByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrLogGroupName], rs.Primary.Attributes[names.AttrName])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccSubscriptionFilterImportStateIDFunc(resourceName string) resource.ImportStateIdFunc {
	return acctest.AttrsImportStateIdFunc(resourceName, "|", names.AttrLogGroupName, names.AttrName)
}

func testAccCheckSubscriptionFilterManyExists(ctx context.Context, t *testing.T, basename string, n int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for i := range n {
			n := fmt.Sprintf("%s.%d", basename, i)
			var v types.SubscriptionFilter

			err := testAccCheckSubscriptionFilterExists(ctx, t, n, &v)(s)

			if err != nil {
				return err
			}
		}

		return nil
	}
}

func testAccSubscriptionFilterConfig_kinesisDataFirehoseBase(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

data "aws_region" "current" {}

resource "aws_cloudwatch_log_group" "test" {
  name              = %[1]q
  retention_in_days = 1
}

resource "aws_iam_role" "firehose" {
  name = "%[1]s-firehose"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "firehose.amazonaws.com"
      },
      "Action": "sts:AssumeRole",
      "Condition": {
        "StringEquals": {
          "sts:ExternalId": "${data.aws_caller_identity.current.account_id}"
        }
      }
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "firehose" {
  role = aws_iam_role.firehose.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Action": [
        "s3:AbortMultipartUpload",
        "s3:GetBucketLocation",
        "s3:GetObject",
        "s3:ListBucket",
        "s3:ListBucketMultipartUploads",
        "s3:PutObject"
      ],
      "Resource": [
        "${aws_s3_bucket.test.arn}",
        "${aws_s3_bucket.test.arn}/*"
      ]
    }
  ]
}
EOF
}

resource "aws_iam_role" "cloudwatchlogs" {
  name = "%[1]s-cloudwatchlogs"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "logs.${data.aws_region.current.region}.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "cloudwatchlogs" {
  role = aws_iam_role.cloudwatchlogs.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": "firehose:*",
      "Resource": "arn:${data.aws_partition.current.partition}:firehose:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}:*"
    },
    {
      "Effect": "Allow",
      "Action": "iam:PassRole",
      "Resource": "${aws_iam_role.cloudwatchlogs.arn}"
    }
  ]
}
EOF
}

resource "aws_kinesis_firehose_delivery_stream" "test" {
  destination = "extended_s3"
  name        = %[1]q

  extended_s3_configuration {
    bucket_arn = aws_s3_bucket.test.arn
    role_arn   = aws_iam_role.firehose.arn
  }
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}
`, rName)
}

func testAccSubscriptionFilterConfig_kinesisStreamBase(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_cloudwatch_log_group" "test" {
  name              = %[1]q
  retention_in_days = 1
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "logs.${data.aws_region.current.region}.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  role = aws_iam_role.test.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": "kinesis:PutRecord",
      "Resource": "${aws_kinesis_stream.test.arn}"
    },
    {
      "Effect": "Allow",
      "Action": "iam:PassRole",
      "Resource": "${aws_iam_role.test.arn}"
    }
  ]
}
EOF
}

resource "aws_kinesis_stream" "test" {
  name        = %[1]q
  shard_count = 1
}
`, rName)
}

func testAccSubscriptionFilterConfig_lambdaBase(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_cloudwatch_log_group" "test" {
  name              = %[1]q
  retention_in_days = 1
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "test" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
  role       = aws_iam_role.test.name
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.test.arn
  runtime       = "nodejs24.x"
  handler       = "exports.handler"
}

resource "aws_lambda_permission" "test" {
  statement_id  = "AllowExecutionFromCloudWatchLogs"
  action        = "lambda:*"
  function_name = aws_lambda_function.test.arn
  principal     = "logs.amazonaws.com"
}
`, rName)
}

func testAccSubscriptionFilterConfig_lambdaMany(rName string, n int) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_cloudwatch_log_group" "test" {
  name              = %[1]q
  retention_in_days = 1
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "test" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
  role       = aws_iam_role.test.name
}

resource "aws_lambda_function" "test" {
  count = %[2]d

  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%[1]s-${count.index}"
  role          = aws_iam_role.test.arn
  runtime       = "nodejs24.x"
  handler       = "exports.handler"
}

resource "aws_lambda_permission" "test" {
  count = 2

  statement_id  = "AllowExecutionFromCloudWatchLogs"
  action        = "lambda:*"
  function_name = aws_lambda_function.test[count.index].arn
  principal     = "logs.amazonaws.com"
}
`, rName, n)
}

func testAccSubscriptionFilterConfig_kinesisDataFirehoseCrossAccountBase_initProviders() string {
	return acctest.ConfigCompose(acctest.ConfigAlternateAccountProvider(), `
data "aws_caller_identity" "destination" {
  provider = awsalternate
}

data "aws_caller_identity" "source" {}
`)
}

func testAccSubscriptionFilterConfig_kinesisDataFirehoseCrossAccountBase(rName string) string {
	return acctest.ConfigCompose(testAccSubscriptionFilterConfig_kinesisDataFirehoseCrossAccountBase_initProviders(), fmt.Sprintf(`
data "aws_region" "destination" {
  provider = awsalternate
}

data "aws_partition" "destination" {
  provider = awsalternate
}

#
# Creation of firehose delivery stream
#
resource "aws_s3_bucket" "log_collector" {
  provider = awsalternate

  bucket = %[1]q
}

resource "aws_iam_role" "firehose_put_data_into_bucket" {
  provider = awsalternate

  name = "%[1]s-AllowFirehoseToPutDataIntoBucket"
  assume_role_policy = jsonencode({
    "Version" = "2012-10-17",
    "Statement" = [{
      "Effect" = "Allow",
      "Principal" = {
        "Service" = "firehose.amazonaws.com"
      },
      "Action" = "sts:AssumeRole",
      "Condition" = {
        "StringEquals" = {
          "sts:ExternalId" = data.aws_caller_identity.destination.account_id
        }
      }
    }]
  })
}

resource "aws_iam_role_policy" "permission_for_firehose" {
  provider = awsalternate

  role = aws_iam_role.firehose_put_data_into_bucket.id
  policy = jsonencode({
    "Statement" = [{
      "Effect" = "Allow",
      "Action" = [
        "s3:PutObject",
        "s3:PutObjectAcl",
        "s3:ListBucket"
      ],
      "Resource" = [
        "arn:${data.aws_partition.destination.partition}:s3:::%[1]s",
        "arn:${data.aws_partition.destination.partition}:s3:::%[1]s/*"
      ]
    }]
  })
}

resource "aws_kinesis_firehose_delivery_stream" "extended_s3_stream" {
  provider = awsalternate

  name        = "%[1]s-tf-firehose-s3-test-stream"
  destination = "extended_s3"

  extended_s3_configuration {
    role_arn   = aws_iam_role.firehose_put_data_into_bucket.arn
    bucket_arn = aws_s3_bucket.log_collector.arn
  }
}

#
# Creation of a destination with org ID in destination access policy
#
resource "aws_iam_role" "cloudwatch_to_firehose" {
  provider = awsalternate

  name = "%[1]s-CWLtoKinesisFirehoseRole"

  assume_role_policy = jsonencode({
    "Version" = "2012-10-17",
    "Statement" = [{
      "Effect" = "Allow",
      "Principal" = {
        "Service" = "logs.${data.aws_region.destination.region}.amazonaws.com"
      },
      "Action" = "sts:AssumeRole",
      "Condition" = {
        "StringLike" = {
          "aws:SourceArn" = [
            "arn:${data.aws_partition.destination.partition}:logs:${data.aws_region.destination.region}:${data.aws_caller_identity.source.account_id}:*",
            "arn:${data.aws_partition.destination.partition}:logs:${data.aws_region.destination.region}:${data.aws_caller_identity.destination.account_id}:*"
          ]
        }
      }
    }]
  })
}

resource "aws_iam_role_policy" "cloudwatch_allow_firehose" {
  provider = awsalternate

  role = aws_iam_role.cloudwatch_to_firehose.id
  policy = jsonencode({
    "Statement" = [
      {
        "Effect"   = "Allow",
        "Action"   = ["firehose:*"],
        "Resource" = ["arn:${data.aws_partition.destination.partition}:firehose:${data.aws_region.destination.region}:${data.aws_caller_identity.destination.account_id}:*"]
      }
    ]
  })
}

resource "aws_cloudwatch_log_destination" "firehose" {
  provider = awsalternate

  name       = "%[1]s-testFirehoseDestination"
  role_arn   = aws_iam_role.cloudwatch_to_firehose.arn
  target_arn = aws_kinesis_firehose_delivery_stream.extended_s3_stream.arn
}

data "aws_organizations_organization" "current" {
  provider = awsalternate
}

resource "aws_cloudwatch_log_destination_policy" "firehose_destination_policy" {
  provider = awsalternate

  destination_name = aws_cloudwatch_log_destination.firehose.name

  access_policy = jsonencode({
    "Version" = "2012-10-17",
    "Statement" = [
      {
        "Sid"       = "",
        "Effect"    = "Allow",
        "Principal" = "*",
        "Action"    = "logs:PutSubscriptionFilter",
        "Resource"  = aws_cloudwatch_log_destination.firehose.arn,
        "Condition" = {
          "StringEquals" = {
            "aws:PrincipalOrgID" = [
              data.aws_organizations_organization.current.id
            ]
          }
        }
      }
    ]
  })
}
`, rName))
}

func testAccSubscriptionFilterConfig_destinationARNKinesisDataFirehose(rName string) string {
	return acctest.ConfigCompose(testAccSubscriptionFilterConfig_kinesisDataFirehoseBase(rName), fmt.Sprintf(`
resource "aws_cloudwatch_log_subscription_filter" "test" {
  destination_arn = aws_kinesis_firehose_delivery_stream.test.arn
  filter_pattern  = "logtype test"
  log_group_name  = aws_cloudwatch_log_group.test.name
  name            = %[1]q
  role_arn        = aws_iam_role.cloudwatchlogs.arn
}
`, rName))
}

func testAccSubscriptionFilterConfig_destinationARNKinesisStream(rName string) string {
	return acctest.ConfigCompose(testAccSubscriptionFilterConfig_kinesisStreamBase(rName), fmt.Sprintf(`
resource "aws_cloudwatch_log_subscription_filter" "test" {
  destination_arn = aws_kinesis_stream.test.arn
  filter_pattern  = "logtype test"
  log_group_name  = aws_cloudwatch_log_group.test.name
  name            = %[1]q
  role_arn        = aws_iam_role.test.arn
}
`, rName))
}

func testAccSubscriptionFilterConfig_destinationARNLambda(rName string) string {
	return acctest.ConfigCompose(testAccSubscriptionFilterConfig_lambdaBase(rName), fmt.Sprintf(`
resource "aws_cloudwatch_log_subscription_filter" "test" {
  destination_arn = aws_lambda_function.test.arn
  filter_pattern  = "logtype test"
  log_group_name  = aws_cloudwatch_log_group.test.name
  name            = %[1]q
}
`, rName))
}

func testAccSubscriptionFilterConfig_distribution(rName, distribution string) string {
	return acctest.ConfigCompose(testAccSubscriptionFilterConfig_lambdaBase(rName), fmt.Sprintf(`
resource "aws_cloudwatch_log_subscription_filter" "test" {
  destination_arn = aws_lambda_function.test.arn
  distribution    = %[2]q
  filter_pattern  = "logtype test"
  log_group_name  = aws_cloudwatch_log_group.test.name
  name            = %[1]q
}
`, rName, distribution))
}

func testAccSubscriptionFilterConfig_destinationARNLambdaMany(rName string, n int) string {
	return acctest.ConfigCompose(testAccSubscriptionFilterConfig_lambdaMany(rName, n), fmt.Sprintf(`
resource "aws_cloudwatch_log_subscription_filter" "test" {
  count = %[2]d

  destination_arn = aws_lambda_function.test[count.index].arn
  filter_pattern  = "logtype test"
  log_group_name  = aws_cloudwatch_log_group.test.name
  name            = "%[1]s-${count.index}"
}
`, rName, n))
}

func testAccSubscriptionFilterConfig_emitSystemFields(rName, emitSystemFieldsStr string) string {
	return acctest.ConfigCompose(testAccSubscriptionFilterConfig_lambdaBase(rName), fmt.Sprintf(`
resource "aws_cloudwatch_log_subscription_filter" "test" {
  destination_arn    = aws_lambda_function.test.arn
  filter_pattern     = "logtype test"
  log_group_name     = aws_cloudwatch_log_group.test.name
  name               = %[1]q
  emit_system_fields = %[2]s
}
`, rName, emitSystemFieldsStr))
}

func testAccSubscriptionFilterConfig_roleARN1(rName string) string {
	return acctest.ConfigCompose(testAccSubscriptionFilterConfig_kinesisStreamBase(rName), fmt.Sprintf(`
resource "aws_cloudwatch_log_subscription_filter" "test" {
  destination_arn = aws_kinesis_stream.test.arn
  filter_pattern  = "logtype test"
  log_group_name  = aws_cloudwatch_log_group.test.name
  name            = %[1]q
  role_arn        = aws_iam_role.test.arn
}
`, rName))
}

func testAccSubscriptionFilterConfig_roleARN2(rName string) string {
	return acctest.ConfigCompose(testAccSubscriptionFilterConfig_kinesisStreamBase(rName), fmt.Sprintf(`
resource "aws_iam_role" "test2" {
  name = "%[1]s-2"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "logs.${data.aws_region.current.region}.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test2" {
  role = aws_iam_role.test2.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": "kinesis:PutRecord",
      "Resource": "${aws_kinesis_stream.test.arn}"
    },
    {
      "Effect": "Allow",
      "Action": "iam:PassRole",
      "Resource": "${aws_iam_role.test2.arn}"
    }
  ]
}
EOF
}

resource "aws_cloudwatch_log_subscription_filter" "test" {
  destination_arn = aws_kinesis_stream.test.arn
  filter_pattern  = "logtype test"
  log_group_name  = aws_cloudwatch_log_group.test.name
  name            = %[1]q
  role_arn        = aws_iam_role.test2.arn
}
`, rName))
}

func testAccSubscriptionFilterConfig_applyOnTransformed(rName string, applyOnTransformedLogs bool) string {
	return acctest.ConfigCompose(testAccSubscriptionFilterConfig_lambdaBase(rName), fmt.Sprintf(`
resource "aws_cloudwatch_log_transformer" "test" {
  log_group_arn = aws_cloudwatch_log_group.test.arn

  transformer_config {
    parse_json {}
  }
}

resource "aws_cloudwatch_log_subscription_filter" "test" {
  destination_arn = aws_lambda_function.test.arn
  filter_pattern  = "logtype test"
  log_group_name  = aws_cloudwatch_log_group.test.name
  name            = %[1]q

  apply_on_transformed_logs = %[2]t
}
`, rName, applyOnTransformedLogs))
}

func testAccSubscriptionFilterConfig_destinationARNKinesisDataFirehoseCrossAccount(rName string) string {
	return acctest.ConfigCompose(testAccSubscriptionFilterConfig_kinesisDataFirehoseCrossAccountBase(rName), fmt.Sprintf(`
data "aws_region" "source" {}

data "aws_partition" "source" {}

resource "aws_cloudwatch_log_group" "source" {
  name              = "%[1]s-testSourceCWLGroup"
  retention_in_days = 1
}

resource "aws_iam_role" "cloudwatch_subscription_filter" {
  name = "%[1]s-CWLtoSubscriptionFilterRole"
  assume_role_policy = jsonencode({
    "Version" = "2012-10-17",
    "Statement" = [
      {
        "Effect" = "Allow",
        "Principal" = {
          "Service" = "logs.${data.aws_region.source.region}.amazonaws.com"
        },
        "Action" = "sts:AssumeRole"
      }
    ]
  })
}

resource "aws_iam_role_policy" "cloudwatch_put_log_events" {
  role = aws_iam_role.cloudwatch_subscription_filter.id
  policy = jsonencode({
    "Statement" = [
      {
        "Effect"   = "Allow",
        "Action"   = "logs:PutLogEvents",
        "Resource" = "arn:${data.aws_partition.source.partition}:logs:${data.aws_region.source.region}:${data.aws_caller_identity.source.account_id}:log-group:${aws_cloudwatch_log_group.source.name}:*"
      }
    ]
  })
}

resource "aws_cloudwatch_log_subscription_filter" "firehose" {
  name            = "%[1]s-firehoseTestSubFilter"
  destination_arn = aws_cloudwatch_log_destination.firehose.arn
  log_group_name  = aws_cloudwatch_log_group.source.name
  filter_pattern  = ""
  role_arn        = aws_iam_role.cloudwatch_subscription_filter.arn
}
`, rName))
}
