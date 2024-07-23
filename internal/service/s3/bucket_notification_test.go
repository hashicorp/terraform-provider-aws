// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfs3 "github.com/hashicorp/terraform-provider-aws/internal/service/s3"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3BucketNotification_eventbridge(t *testing.T) {
	ctx := acctest.Context(t)
	var v s3.GetBucketNotificationConfigurationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_notification.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketNotificationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketNotificationConfig_eventBridge(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketNotificationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "eventbridge", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "lambda_function.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "queue.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "topic.#", acctest.Ct0),
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

func TestAccS3BucketNotification_lambdaFunction(t *testing.T) {
	ctx := acctest.Context(t)
	var v s3.GetBucketNotificationConfigurationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_notification.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketNotificationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketNotificationConfig_lambdaFunction(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketNotificationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "eventbridge", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "lambda_function.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "lambda_function.0.events.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "lambda_function.0.filter_prefix", "tf-acc-test/"),
					resource.TestCheckResourceAttr(resourceName, "lambda_function.0.filter_suffix", ".png"),
					resource.TestCheckResourceAttr(resourceName, "queue.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "topic.#", acctest.Ct0),
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

func TestAccS3BucketNotification_LambdaFunctionLambdaFunctionARN_alias(t *testing.T) {
	ctx := acctest.Context(t)
	var v s3.GetBucketNotificationConfigurationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_notification.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketNotificationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketNotificationConfig_lambdaFunctionLambdaFunctionARNAlias(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketNotificationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "eventbridge", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "lambda_function.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "lambda_function.0.events.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "lambda_function.0.filter_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "lambda_function.0.filter_suffix", ""),
					resource.TestCheckResourceAttr(resourceName, "queue.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "topic.#", acctest.Ct0),
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

func TestAccS3BucketNotification_queue(t *testing.T) {
	ctx := acctest.Context(t)
	var v s3.GetBucketNotificationConfigurationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_notification.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketNotificationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketNotificationConfig_queue(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketNotificationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "eventbridge", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "lambda_function.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "queue.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "queue.0.events.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "queue.0.filter_prefix", "tf-acc-test/"),
					resource.TestCheckResourceAttr(resourceName, "queue.0.filter_suffix", ".mp4"),
					resource.TestCheckResourceAttr(resourceName, "topic.#", acctest.Ct0),
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

func TestAccS3BucketNotification_topic(t *testing.T) {
	ctx := acctest.Context(t)
	var v s3.GetBucketNotificationConfigurationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_notification.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketNotificationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketNotificationConfig_topic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketNotificationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "eventbridge", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "lambda_function.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "queue.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "topic.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "topic.0.events.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "topic.0.filter_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "topic.0.filter_suffix", ""),
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

func TestAccS3BucketNotification_Topic_multiple(t *testing.T) {
	ctx := acctest.Context(t)
	var v s3.GetBucketNotificationConfigurationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_notification.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketNotificationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketNotificationConfig_topicMultiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketNotificationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "eventbridge", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "lambda_function.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "queue.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "topic.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "topic.0.events.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "topic.0.filter_prefix", "tf-acc-test/"),
					resource.TestCheckResourceAttr(resourceName, "topic.0.filter_suffix", ".txt"),
					resource.TestCheckResourceAttr(resourceName, "topic.1.events.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "topic.1.filter_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "topic.1.filter_suffix", ".log"),
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

func TestAccS3BucketNotification_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v s3.GetBucketNotificationConfigurationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_notification.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketNotificationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketNotificationConfig_topic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketNotificationExists(ctx, resourceName, &v),
				),
			},
			{
				Config: testAccBucketNotificationConfig_queue(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketNotificationExists(ctx, resourceName, &v),
				),
			},
		},
	})
}

func TestAccS3BucketNotification_directoryBucket(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketNotificationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccBucketNotificationConfig_directoryBucket(rName),
				ExpectError: regexache.MustCompile(`directory buckets are not supported`),
			},
		},
	})
}

func testAccCheckBucketNotificationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3_bucket_notification" {
				continue
			}

			_, err := tfs3.FindBucketNotificationConfiguration(ctx, conn, rs.Primary.ID, "")

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("S3 Bucket Notification %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckBucketNotificationExists(ctx context.Context, n string, v *s3.GetBucketNotificationConfigurationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Client(ctx)

		output, err := tfs3.FindBucketNotificationConfiguration(ctx, conn, rs.Primary.ID, "")

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccBucketNotificationConfig_eventBridge(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_public_access_block" "test" {
  bucket = aws_s3_bucket.test.id

  block_public_acls       = false
  block_public_policy     = false
  ignore_public_acls      = false
  restrict_public_buckets = false
}

resource "aws_s3_bucket_ownership_controls" "test" {
  bucket = aws_s3_bucket.test.id
  rule {
    object_ownership = "BucketOwnerPreferred"
  }
}

resource "aws_s3_bucket_acl" "test" {
  depends_on = [
    aws_s3_bucket_public_access_block.test,
    aws_s3_bucket_ownership_controls.test,
  ]

  bucket = aws_s3_bucket.test.id
  acl    = "public-read"
}

resource "aws_s3_bucket_notification" "test" {
  bucket = aws_s3_bucket.test.id

  eventbridge = true
}
`, rName)
}

func testAccBucketNotificationConfig_topicMultiple(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_sns_topic" "test" {
  name = %[1]q

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": { "Service": "s3.${data.aws_partition.current.dns_suffix}" },
      "Action": "SNS:Publish",
      "Resource": "arn:${data.aws_partition.current.partition}:sns:*:*:%[1]s",
      "Condition": {
        "ArnLike": {
          "aws:SourceArn": "${aws_s3_bucket.test.arn}"
        }
      }
    }
  ]
}
POLICY
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_public_access_block" "test" {
  bucket = aws_s3_bucket.test.id

  block_public_acls       = false
  block_public_policy     = false
  ignore_public_acls      = false
  restrict_public_buckets = false
}

resource "aws_s3_bucket_ownership_controls" "test" {
  bucket = aws_s3_bucket.test.id
  rule {
    object_ownership = "BucketOwnerPreferred"
  }
}

resource "aws_s3_bucket_acl" "test" {
  depends_on = [
    aws_s3_bucket_public_access_block.test,
    aws_s3_bucket_ownership_controls.test,
  ]

  bucket = aws_s3_bucket.test.id
  acl    = "public-read"
}

resource "aws_s3_bucket_notification" "test" {
  bucket = aws_s3_bucket.test.id

  topic {
    id        = "notification-sns1"
    topic_arn = aws_sns_topic.test.arn

    events = [
      "s3:ObjectCreated:*",
      "s3:ObjectRemoved:Delete",
    ]

    filter_prefix = "tf-acc-test/"
    filter_suffix = ".txt"
  }

  topic {
    id        = "notification-sns2"
    topic_arn = aws_sns_topic.test.arn

    events = [
      "s3:ObjectCreated:*",
      "s3:ObjectRemoved:Delete",
    ]

    filter_suffix = ".log"
  }
}
`, rName)
}

func testAccBucketNotificationConfig_queue(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_sqs_queue" "test" {
  name = %[1]q

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": "*",
      "Action": "sqs:SendMessage",
      "Resource": "arn:${data.aws_partition.current.partition}:sqs:*:*:%[1]s",
      "Condition": {
        "ArnEquals": {
          "aws:SourceArn": "${aws_s3_bucket.test.arn}"
        }
      }
    }
  ]
}
POLICY
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_public_access_block" "test" {
  bucket = aws_s3_bucket.test.id

  block_public_acls       = false
  block_public_policy     = false
  ignore_public_acls      = false
  restrict_public_buckets = false
}

resource "aws_s3_bucket_ownership_controls" "test" {
  bucket = aws_s3_bucket.test.id
  rule {
    object_ownership = "BucketOwnerPreferred"
  }
}

resource "aws_s3_bucket_acl" "test" {
  depends_on = [
    aws_s3_bucket_public_access_block.test,
    aws_s3_bucket_ownership_controls.test,
  ]

  bucket = aws_s3_bucket.test.id
  acl    = "public-read"
}

resource "aws_s3_bucket_notification" "test" {
  bucket = aws_s3_bucket.test.id

  queue {
    id        = "notification-sqs"
    queue_arn = aws_sqs_queue.test.arn

    events = [
      "s3:ObjectCreated:*",
      "s3:ObjectRemoved:Delete",
    ]

    filter_prefix = "tf-acc-test/"
    filter_suffix = ".mp4"
  }
}
`, rName)
}

func testAccBucketNotificationConfig_lambdaFunction(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_lambda_permission" "test" {
  statement_id  = "AllowExecutionFromS3Bucket"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.test.arn
  principal     = "s3.${data.aws_partition.current.dns_suffix}"
  source_arn    = aws_s3_bucket.test.arn
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.test.arn
  handler       = "exports.example"
  runtime       = "nodejs16.x"
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_public_access_block" "test" {
  bucket = aws_s3_bucket.test.id

  block_public_acls       = false
  block_public_policy     = false
  ignore_public_acls      = false
  restrict_public_buckets = false
}

resource "aws_s3_bucket_ownership_controls" "test" {
  bucket = aws_s3_bucket.test.id
  rule {
    object_ownership = "BucketOwnerPreferred"
  }
}

resource "aws_s3_bucket_acl" "test" {
  depends_on = [
    aws_s3_bucket_public_access_block.test,
    aws_s3_bucket_ownership_controls.test,
  ]

  bucket = aws_s3_bucket.test.id
  acl    = "public-read"
}

resource "aws_s3_bucket_notification" "test" {
  bucket = aws_s3_bucket.test.id

  lambda_function {
    id                  = "notification-lambda"
    lambda_function_arn = aws_lambda_function.test.arn

    events = [
      "s3:ObjectCreated:*",
      "s3:ObjectRemoved:Delete",
    ]

    filter_prefix = "tf-acc-test/"
    filter_suffix = ".png"
  }
}
`, rName)
}

func testAccBucketNotificationConfig_lambdaFunctionLambdaFunctionARNAlias(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow"
    }
  ]
}
EOF

  name = %[1]q
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  handler       = "exports.example"
  role          = aws_iam_role.test.arn
  runtime       = "nodejs16.x"
}

resource "aws_lambda_alias" "test" {
  function_name    = aws_lambda_function.test.arn
  function_version = "$LATEST"
  name             = "testalias"
}

resource "aws_lambda_permission" "test" {
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.test.arn
  principal     = "s3.${data.aws_partition.current.dns_suffix}"
  qualifier     = aws_lambda_alias.test.name
  source_arn    = aws_s3_bucket.test.arn
  statement_id  = "AllowExecutionFromS3Bucket"
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_notification" "test" {
  bucket = aws_s3_bucket.test.id

  lambda_function {
    events              = ["s3:ObjectCreated:*"]
    id                  = "test"
    lambda_function_arn = aws_lambda_alias.test.arn
  }
}
`, rName)
}

func testAccBucketNotificationConfig_topic(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_sns_topic" "test" {
  name = %[1]q

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "s3.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "SNS:Publish",
      "Resource": "arn:${data.aws_partition.current.partition}:sns:*:*:%[1]s",
      "Condition": {
        "ArnLike": {
          "aws:SourceArn": "${aws_s3_bucket.test.arn}"
        }
      }
    }
  ]
}
POLICY
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_public_access_block" "test" {
  bucket = aws_s3_bucket.test.id

  block_public_acls       = false
  block_public_policy     = false
  ignore_public_acls      = false
  restrict_public_buckets = false
}

resource "aws_s3_bucket_ownership_controls" "test" {
  bucket = aws_s3_bucket.test.id
  rule {
    object_ownership = "BucketOwnerPreferred"
  }
}

resource "aws_s3_bucket_acl" "test" {
  depends_on = [
    aws_s3_bucket_public_access_block.test,
    aws_s3_bucket_ownership_controls.test,
  ]

  bucket = aws_s3_bucket.test.id
  acl    = "public-read"
}

resource "aws_s3_bucket_notification" "test" {
  bucket = aws_s3_bucket.test.id

  topic {
    id        = "notification-sns1"
    topic_arn = aws_sns_topic.test.arn

    events = [
      "s3:ObjectCreated:*",
      "s3:ObjectRemoved:Delete",
    ]
  }
}
`, rName)
}

func testAccBucketNotificationConfig_directoryBucket(rName string) string {
	return acctest.ConfigCompose(testAccDirectoryBucketConfig_base(rName), `
resource "aws_s3_directory_bucket" "test" {
  bucket = local.bucket

  location {
    name = local.location_name
  }
}

resource "aws_s3_bucket_notification" "test" {
  bucket = aws_s3_directory_bucket.test.bucket

  eventbridge = true
}
`)
}
