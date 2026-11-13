// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3BucketNotificationDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_s3_bucket_notification.test"
	lambdaResourceName := "aws_lambda_function.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketNotificationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketNotificationDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "eventbridge", acctest.CtFalse),
					resource.TestCheckResourceAttr(dataSourceName, "lambda_function.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "lambda_function.0.id", "notification-lambda"),
					resource.TestCheckResourceAttrPair(dataSourceName, "lambda_function.0.lambda_function_arn", lambdaResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(dataSourceName, "lambda_function.0.events.#", "2"),
					resource.TestCheckResourceAttr(dataSourceName, "lambda_function.0.filter_prefix", "tf-acc-test/"),
					resource.TestCheckResourceAttr(dataSourceName, "lambda_function.0.filter_suffix", ".png"),
					resource.TestCheckResourceAttr(dataSourceName, "queue.#", "0"),
					resource.TestCheckResourceAttr(dataSourceName, "topic.#", "0"),
				),
			},
		},
	})
}

func TestAccS3BucketNotificationDataSource_eventbridge(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_s3_bucket_notification.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketNotificationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketNotificationDataSourceConfig_eventbridge(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "eventbridge", acctest.CtTrue),
					resource.TestCheckResourceAttr(dataSourceName, "lambda_function.#", "0"),
					resource.TestCheckResourceAttr(dataSourceName, "queue.#", "0"),
					resource.TestCheckResourceAttr(dataSourceName, "topic.#", "0"),
				),
			},
		},
	})
}

func TestAccS3BucketNotificationDataSource_readAndReemit(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_s3_bucket_notification.from_source"
	mirrorResourceName := "aws_s3_bucket_notification.mirror"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketNotificationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketNotificationDataSourceConfig_readAndReemit(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Data source surfaces the source bucket's config.
					resource.TestCheckResourceAttr(dataSourceName, "eventbridge", acctest.CtTrue),
					resource.TestCheckResourceAttr(dataSourceName, "lambda_function.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "lambda_function.0.id", "team-a-rule"),
					resource.TestCheckResourceAttr(dataSourceName, "lambda_function.0.filter_prefix", "tf-acc-test/"),
					resource.TestCheckResourceAttr(dataSourceName, "lambda_function.0.filter_suffix", ".png"),
					// Mirror bucket faithfully reproduces it through dynamic.
					resource.TestCheckResourceAttr(mirrorResourceName, "eventbridge", acctest.CtTrue),
					resource.TestCheckResourceAttr(mirrorResourceName, "lambda_function.#", "1"),
					resource.TestCheckResourceAttr(mirrorResourceName, "lambda_function.0.id", "team-a-rule"),
					resource.TestCheckResourceAttr(mirrorResourceName, "lambda_function.0.filter_prefix", "tf-acc-test/"),
					resource.TestCheckResourceAttr(mirrorResourceName, "lambda_function.0.filter_suffix", ".png"),
				),
			},
		},
	})
}

func testAccBucketNotificationDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccBucketNotificationConfig_lambdaFunction(rName),
		`
data "aws_s3_bucket_notification" "test" {
  bucket = aws_s3_bucket_notification.test.bucket
}
`,
	)
}

func testAccBucketNotificationDataSourceConfig_eventbridge(rName string) string {
	return acctest.ConfigCompose(
		testAccBucketNotificationConfig_eventBridge(rName),
		`
data "aws_s3_bucket_notification" "test" {
  bucket = aws_s3_bucket_notification.test.bucket
}
`,
	)
}

func testAccBucketNotificationDataSourceConfig_readAndReemit(rName string) string {
	return fmt.Sprintf(`
data "aws_service_principal" "s3" {
  service_name = "s3"
}

data "aws_service_principal" "lambda" {
  service_name = "lambda"
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Principal = {
        Service = data.aws_service_principal.lambda.name
      }
      Action = "sts:AssumeRole"
    }]
  })
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.test.arn
  handler       = "exports.example"
  runtime       = "nodejs24.x"
}

resource "aws_s3_bucket" "source" {
  bucket = "%[1]s-src"
}

resource "aws_s3_bucket" "mirror" {
  bucket = "%[1]s-mir"
}

resource "aws_lambda_permission" "source" {
  statement_id  = "AllowExecutionFromS3Source"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.test.arn
  principal     = data.aws_service_principal.s3.name
  source_arn    = aws_s3_bucket.source.arn
}

resource "aws_lambda_permission" "mirror" {
  statement_id  = "AllowExecutionFromS3Mirror"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.test.arn
  principal     = data.aws_service_principal.s3.name
  source_arn    = aws_s3_bucket.mirror.arn
}

resource "aws_s3_bucket_notification" "source" {
  bucket      = aws_s3_bucket.source.id
  eventbridge = true

  lambda_function {
    id                  = "team-a-rule"
    lambda_function_arn = aws_lambda_function.test.arn
    events              = ["s3:ObjectCreated:*"]
    filter_prefix       = "tf-acc-test/"
    filter_suffix       = ".png"
  }

  depends_on = [aws_lambda_permission.source]
}

data "aws_s3_bucket_notification" "from_source" {
  bucket = aws_s3_bucket_notification.source.bucket
}

resource "aws_s3_bucket_notification" "mirror" {
  bucket      = aws_s3_bucket.mirror.id
  eventbridge = data.aws_s3_bucket_notification.from_source.eventbridge

  dynamic "lambda_function" {
    for_each = data.aws_s3_bucket_notification.from_source.lambda_function
    content {
      id                  = lambda_function.value.id
      lambda_function_arn = lambda_function.value.lambda_function_arn
      events              = lambda_function.value.events
      filter_prefix       = lambda_function.value.filter_prefix
      filter_suffix       = lambda_function.value.filter_suffix
    }
  }

  depends_on = [aws_lambda_permission.mirror]
}
`, rName)
}
