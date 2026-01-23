package s3_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/s3"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccS3BucketNotificationDataSource_lambdaFunction(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	resourceName := "data.aws_s3_bucket_notification.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, s3.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketNotificationDataSourceConfig_lambdaFunction(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "lambda_function_configurations.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_function_configurations.0.lambda_function_arn", "aws_lambda_function.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "lambda_function_configurations.0.id", "notification-lambda"),
					resource.TestCheckResourceAttr(resourceName, "lambda_function_configurations.0.filter_prefix", "tf-acc-test/"),
					resource.TestCheckResourceAttr(resourceName, "lambda_function_configurations.0.filter_suffix", ".png"),
					resource.TestCheckResourceAttr(resourceName, "lambda_function_configurations.0.events.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "lambda_function_configurations.0.events.0", "s3:ObjectCreated:*"),
					resource.TestCheckResourceAttr(resourceName, "lambda_function_configurations.0.events.1", "s3:ObjectRemoved:Delete"),
				),
			},
		},
	})
}

func testAccBucketNotificationDataSourceConfig_lambdaFunction(rName string) string {
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

data "aws_s3_bucket_notification" "test" {
  bucket = aws_s3_bucket.test.id
}
`, rName)
}

func TestAccS3BucketNotificationDataSource_queue(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	resourceName := "data.aws_s3_bucket_notification.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, s3.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketNotificationDataSourceConfig_queue(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "queue_configurations.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "queue_configurations.0.queue_arn", "aws_lambda_function.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "queue_configurations.0.id", "notification-sqs"),
					resource.TestCheckResourceAttr(resourceName, "queue_configurations.0.filter_prefix", "tf-acc-test/"),
					resource.TestCheckResourceAttr(resourceName, "queue_configurations.0.filter_suffix", ".mp4"),
					resource.TestCheckResourceAttr(resourceName, "queue_configurations.0.events.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "queue_configurations.0.events.0", "s3:ObjectCreated:*"),
					resource.TestCheckResourceAttr(resourceName, "queue_configurations.0.events.1", "s3:ObjectRemoved:Delete"),
				),
			},
		},
	})
}

func testAccBucketNotificationDataSourceConfig_queue(rName string) string {
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

data "aws_s3_bucket_notification" "test" {
  bucket = aws_s3_bucket.test.id
}
`, rName)
}

func TestAccS3BucketNotificationDataSource_topic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	resourceName := "data.aws_s3_bucket_notification.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, s3.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketNotificationDataSourceConfig_topic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "topic_configurations.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "topic_configurations.0.topic_arn", "aws_lambda_function.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "topic_configurations.0.id", "notification-sns1"),
					resource.TestCheckResourceAttr(resourceName, "topic_configurations.0.filter_prefix", "tf-acc-test-topic/"),
					resource.TestCheckResourceAttr(resourceName, "topic_configurations.0.filter_suffix", ".png"),
					resource.TestCheckResourceAttr(resourceName, "topic_configurations.0.events.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "topic_configurations.0.events.0", "s3:ObjectCreated:*"),
					resource.TestCheckResourceAttr(resourceName, "topic_configurations.0.events.1", "s3:ObjectRemoved:Delete"),
				),
			},
		},
	})
}

func testAccBucketNotificationDataSourceConfig_topic(rName string) string {
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

resource "aws_s3_bucket_notification" "test" {
  bucket = aws_s3_bucket.test.id

  topic {
    id        = "notification-sns1"
    topic_arn = aws_sns_topic.test.arn

    events = [
      "s3:ObjectCreated:*",
      "s3:ObjectRemoved:Delete",
    ]

    filter_prefix = "tf-acc-test-topic/"
    filter_suffix = ".png"
  }
}

data "aws_s3_bucket_notification" "test" {
  bucket = aws_s3_bucket.test.id
}
`, rName)
}
