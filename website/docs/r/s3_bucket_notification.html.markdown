---
subcategory: "S3 (Simple Storage)"
layout: "aws"
page_title: "AWS: aws_s3_bucket_notification"
description: |-
  Manages a S3 Bucket Notification Configuration
---

# Resource: aws_s3_bucket_notification

Manages a S3 Bucket Notification Configuration. For additional information, see the [Configuring S3 Event Notifications section in the Amazon S3 Developer Guide](https://docs.aws.amazon.com/AmazonS3/latest/dev/NotificationHowTo.html).

~> **NOTE:** The S3 [`PutBucketNotificationConfiguration`](https://docs.aws.amazon.com/AmazonS3/latest/API/API_PutBucketNotificationConfiguration.html) API is atomic — it replaces the bucket's entire notification configuration on every call. Only one `aws_s3_bucket_notification` resource can manage a bucket; declaring more than one causes a perpetual diff, and applying this resource will overwrite any notifications already on the bucket. To configure multiple destinations on the same bucket, declare them all as nested blocks within a single resource (see [Trigger multiple Lambda functions](#trigger-multiple-lambda-functions) below). To let independent teams or Terraform configurations subscribe to the same bucket without stepping on each other, prefer the [Emit events to EventBridge](#emit-events-to-eventbridge) pattern below. To bring existing notifications under management without losing them, see the [`aws_s3_bucket_notification` data source](../d/s3_bucket_notification.html.markdown#read-existing-notifications-and-re-emit-them).

-> This resource cannot be used with S3 directory buckets.

## Example Usage

### Add notification configuration to SNS Topic

```terraform
data "aws_iam_policy_document" "topic" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["s3.amazonaws.com"]
    }

    actions   = ["SNS:Publish"]
    resources = ["arn:aws:sns:*:*:s3-event-notification-topic"]

    condition {
      test     = "ArnLike"
      variable = "aws:SourceArn"
      values   = [aws_s3_bucket.bucket.arn]
    }
  }
}
resource "aws_sns_topic" "topic" {
  name   = "s3-event-notification-topic"
  policy = data.aws_iam_policy_document.topic.json
}

resource "aws_s3_bucket" "bucket" {
  bucket = "your-bucket-name"
}

resource "aws_s3_bucket_notification" "bucket_notification" {
  bucket = aws_s3_bucket.bucket.id

  topic {
    topic_arn     = aws_sns_topic.topic.arn
    events        = ["s3:ObjectCreated:*"]
    filter_suffix = ".log"
  }
}
```

### Add notification configuration to SQS Queue

```terraform
data "aws_iam_policy_document" "queue" {
  statement {
    effect = "Allow"

    principals {
      type        = "*"
      identifiers = ["*"]
    }

    actions   = ["sqs:SendMessage"]
    resources = ["arn:aws:sqs:*:*:s3-event-notification-queue"]

    condition {
      test     = "ArnEquals"
      variable = "aws:SourceArn"
      values   = [aws_s3_bucket.bucket.arn]
    }
  }
}

resource "aws_sqs_queue" "queue" {
  name   = "s3-event-notification-queue"
  policy = data.aws_iam_policy_document.queue.json
}

resource "aws_s3_bucket" "bucket" {
  bucket = "your-bucket-name"
}

resource "aws_s3_bucket_notification" "bucket_notification" {
  bucket = aws_s3_bucket.bucket.id

  queue {
    queue_arn     = aws_sqs_queue.queue.arn
    events        = ["s3:ObjectCreated:*"]
    filter_suffix = ".log"
  }
}
```

### Add notification configuration to Lambda Function

```terraform
data "aws_iam_policy_document" "assume_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
}

resource "aws_iam_role" "iam_for_lambda" {
  name               = "iam_for_lambda"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_lambda_permission" "allow_bucket" {
  statement_id  = "AllowExecutionFromS3Bucket"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.func.arn
  principal     = "s3.amazonaws.com"
  source_arn    = aws_s3_bucket.bucket.arn
}

resource "aws_lambda_function" "func" {
  filename      = "your-function.zip"
  function_name = "example_lambda_name"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs24.x"
}

resource "aws_s3_bucket" "bucket" {
  bucket = "your-bucket-name"
}

resource "aws_s3_bucket_notification" "bucket_notification" {
  bucket = aws_s3_bucket.bucket.id

  lambda_function {
    lambda_function_arn = aws_lambda_function.func.arn
    events              = ["s3:ObjectCreated:*"]
    filter_prefix       = "AWSLogs/"
    filter_suffix       = ".log"
  }

  depends_on = [aws_lambda_permission.allow_bucket]
}
```

### Trigger multiple Lambda functions

```terraform
data "aws_iam_policy_document" "assume_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
}

resource "aws_iam_role" "iam_for_lambda" {
  name               = "iam_for_lambda"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_lambda_permission" "allow_bucket1" {
  statement_id  = "AllowExecutionFromS3Bucket1"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.func1.arn
  principal     = "s3.amazonaws.com"
  source_arn    = aws_s3_bucket.bucket.arn
}

resource "aws_lambda_function" "func1" {
  filename      = "your-function1.zip"
  function_name = "example_lambda_name1"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs24.x"
}

resource "aws_lambda_permission" "allow_bucket2" {
  statement_id  = "AllowExecutionFromS3Bucket2"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.func2.arn
  principal     = "s3.amazonaws.com"
  source_arn    = aws_s3_bucket.bucket.arn
}

resource "aws_lambda_function" "func2" {
  filename      = "your-function2.zip"
  function_name = "example_lambda_name2"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
}

resource "aws_s3_bucket" "bucket" {
  bucket = "your-bucket-name"
}

resource "aws_s3_bucket_notification" "bucket_notification" {
  bucket = aws_s3_bucket.bucket.id

  lambda_function {
    lambda_function_arn = aws_lambda_function.func1.arn
    events              = ["s3:ObjectCreated:*"]
    filter_prefix       = "AWSLogs/"
    filter_suffix       = ".log"
  }

  lambda_function {
    lambda_function_arn = aws_lambda_function.func2.arn
    events              = ["s3:ObjectCreated:*"]
    filter_prefix       = "OtherLogs/"
    filter_suffix       = ".log"
  }

  depends_on = [
    aws_lambda_permission.allow_bucket1,
    aws_lambda_permission.allow_bucket2,
  ]
}
```

### Add multiple notification configurations to SQS Queue

```terraform
data "aws_iam_policy_document" "queue" {
  statement {
    effect = "Allow"

    principals {
      type        = "*"
      identifiers = ["*"]
    }

    actions   = ["sqs:SendMessage"]
    resources = ["arn:aws:sqs:*:*:s3-event-notification-queue"]

    condition {
      test     = "ArnEquals"
      variable = "aws:SourceArn"
      values   = [aws_s3_bucket.bucket.arn]
    }
  }
}

resource "aws_sqs_queue" "queue" {
  name   = "s3-event-notification-queue"
  policy = data.aws_iam_policy_document.queue.json
}

resource "aws_s3_bucket" "bucket" {
  bucket = "your-bucket-name"
}

resource "aws_s3_bucket_notification" "bucket_notification" {
  bucket = aws_s3_bucket.bucket.id

  queue {
    id            = "image-upload-event"
    queue_arn     = aws_sqs_queue.queue.arn
    events        = ["s3:ObjectCreated:*"]
    filter_prefix = "images/"
  }

  queue {
    id            = "video-upload-event"
    queue_arn     = aws_sqs_queue.queue.arn
    events        = ["s3:ObjectCreated:*"]
    filter_prefix = "videos/"
  }
}
```

For Terraform's [JSON syntax](https://www.terraform.io/docs/configuration/syntax.html), use an array instead of defining the `queue` key twice.

```json
{
	"bucket": "${aws_s3_bucket.bucket.id}",
	"queue": [
		{
			"id": "image-upload-event",
			"queue_arn": "${aws_sqs_queue.queue.arn}",
			"events": ["s3:ObjectCreated:*"],
			"filter_prefix": "images/"
		},
		{
			"id": "video-upload-event",
			"queue_arn": "${aws_sqs_queue.queue.arn}",
			"events": ["s3:ObjectCreated:*"],
			"filter_prefix": "videos/"
		}
	]
}
```

### Emit events to EventBridge

For a bucket shared by multiple independent consumers — different teams, different Terraform configurations, different applications — EventBridge is the recommended pattern. Each consumer subscribes to the bucket through its own [`aws_cloudwatch_event_rule`](cloudwatch_event_rule.html), so they cannot overwrite one another the way notification configurations would.

```terraform
resource "aws_s3_bucket" "shared" {
  bucket = "shared-bucket"
}

resource "aws_s3_bucket_notification" "shared" {
  bucket      = aws_s3_bucket.shared.id
  eventbridge = true
}

# Team A: process new uploads under uploads/
resource "aws_cloudwatch_event_rule" "team_a" {
  name = "team-a-uploads"
  event_pattern = jsonencode({
    source        = ["aws.s3"]
    "detail-type" = ["Object Created"]
    detail = {
      bucket = { name = [aws_s3_bucket.shared.bucket] }
      object = { key = [{ prefix = "uploads/" }] }
    }
  })
}

resource "aws_cloudwatch_event_target" "team_a" {
  rule = aws_cloudwatch_event_rule.team_a.name
  arn  = aws_lambda_function.team_a_processor.arn
}

# Team B: archive deletions under archive/, declared in a separate
# Terraform configuration that knows nothing about Team A.
resource "aws_cloudwatch_event_rule" "team_b" {
  name = "team-b-deletions"
  event_pattern = jsonencode({
    source        = ["aws.s3"]
    "detail-type" = ["Object Deleted"]
    detail = {
      bucket = { name = [aws_s3_bucket.shared.bucket] }
      object = { key = [{ prefix = "archive/" }] }
    }
  })
}

resource "aws_cloudwatch_event_target" "team_b" {
  rule = aws_cloudwatch_event_rule.team_b.name
  arn  = aws_sqs_queue.team_b_archive.arn
}
```

For sharing a bucket between Terraform configurations when EventBridge is not an option, use the [`aws_s3_bucket_notification` data source](../d/s3_bucket_notification.html.markdown#read-existing-notifications-and-re-emit-them) to read existing notifications and re-emit them in your own resource.

## Argument Reference

The following arguments are required:

* `bucket` - (Required) Name of the bucket for notification configuration.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `eventbridge` - (Optional) Whether to enable Amazon EventBridge notifications. Defaults to `false`.
* `lambda_function` - (Optional, Multiple) Used to configure notifications to a Lambda Function. See below.
* `queue` - (Optional) Notification configuration to SQS Queue. See below.
* `topic` - (Optional) Notification configuration to SNS Topic. See below.

### `lambda_function`

* `events` - (Required) [Event](http://docs.aws.amazon.com/AmazonS3/latest/dev/NotificationHowTo.html#notification-how-to-event-types-and-destinations) for which to send notifications.
* `filter_prefix` - (Optional) Object key name prefix.
* `filter_suffix` - (Optional) Object key name suffix.
* `id` - (Optional) Unique identifier for each of the notification configurations.
* `lambda_function_arn` - (Required) Lambda function ARN.

### `queue`

* `events` - (Required) Specifies [event](http://docs.aws.amazon.com/AmazonS3/latest/dev/NotificationHowTo.html#notification-how-to-event-types-and-destinations) for which to send notifications.
* `filter_prefix` - (Optional) Object key name prefix.
* `filter_suffix` - (Optional) Object key name suffix.
* `id` - (Optional) Unique identifier for each of the notification configurations.
* `queue_arn` - (Required) SQS queue ARN.

### `topic`

* `events` - (Required) [Event](http://docs.aws.amazon.com/AmazonS3/latest/dev/NotificationHowTo.html#notification-how-to-event-types-and-destinations) for which to send notifications.
* `filter_prefix` - (Optional) Object key name prefix.
* `filter_suffix` - (Optional) Object key name suffix.
* `id` - (Optional) Unique identifier for each of the notification configurations.
* `topic_arn` - (Required) SNS topic ARN.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_s3_bucket_notification.bucket_notification
  identity = {
    bucket = "bucket-name"
  }
}

resource "aws_s3_bucket_notification" "bucket_notification" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `bucket` (String) Name of the bucket.

#### Optional

* `account_id` (String) Account ID where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import S3 bucket notification using the `bucket`. For example:

```terraform
import {
  to = aws_s3_bucket_notification.bucket_notification
  id = "bucket-name"
}
```

Using `terraform import`, import S3 bucket notification using the `bucket`. For example:

```console
% terraform import aws_s3_bucket_notification.bucket_notification bucket-name
```
