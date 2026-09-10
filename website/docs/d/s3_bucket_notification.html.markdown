---
subcategory: "S3 (Simple Storage)"
layout: "aws"
page_title: "AWS: aws_s3_bucket_notification"
description: |-
  Provides details about the notification configuration of an S3 bucket.
---

# Data Source: aws_s3_bucket_notification

Provides details about the notification configuration of an S3 bucket.

Useful when [`aws_s3_bucket_notification`](../r/s3_bucket_notification.html.markdown) is the right resource but the bucket already has notifications you do not manage. Read the existing notifications with this data source and re-emit them — alongside your own — in a single `aws_s3_bucket_notification` resource. See [issue #501](https://github.com/hashicorp/terraform-provider-aws/issues/501) for the longer story. For sharing a bucket across many independent consumers, enabling [EventBridge](../r/s3_bucket_notification.html.markdown#emit-events-to-eventbridge) on the resource is usually a better fit.

## Example Usage

### Basic Usage

```terraform
data "aws_s3_bucket_notification" "example" {
  bucket = "example-bucket"
}
```

### Conditionally Subscribe via EventBridge

When the bucket forwards events to [Amazon EventBridge](../r/s3_bucket_notification.html.markdown#emit-events-to-eventbridge), independent consumers can subscribe with their own `aws_cloudwatch_event_rule` resources. Use this data source to subscribe only when EventBridge is in fact enabled on the bucket.

```terraform
data "aws_s3_bucket_notification" "shared" {
  bucket = "shared-bucket"
}

resource "aws_cloudwatch_event_rule" "s3_object_created" {
  count       = data.aws_s3_bucket_notification.shared.eventbridge ? 1 : 0
  name        = "shared-bucket-object-created"
  description = "S3 object-created events from the shared bucket."

  event_pattern = jsonencode({
    source        = ["aws.s3"]
    "detail-type" = ["Object Created"]
    detail = {
      bucket = {
        name = [data.aws_s3_bucket_notification.shared.bucket]
      }
    }
  })
}
```

### Read Existing Notifications and Re-emit Them

The S3 [`PutBucketNotificationConfiguration`](https://docs.aws.amazon.com/AmazonS3/latest/API/API_PutBucketNotificationConfiguration.html) API replaces the entire notification configuration on every call, so a single `aws_s3_bucket_notification` resource owns the bucket. To preserve notifications already on the bucket — or to mirror one bucket's configuration onto another — read them with this data source and pass them through `dynamic` blocks. The data source's output shape matches the resource's input shape, so each block forwards directly.

```terraform
data "aws_s3_bucket_notification" "existing" {
  bucket = aws_s3_bucket.example.id
}

resource "aws_s3_bucket_notification" "example" {
  bucket      = aws_s3_bucket.example.id
  eventbridge = data.aws_s3_bucket_notification.existing.eventbridge

  dynamic "lambda_function" {
    for_each = data.aws_s3_bucket_notification.existing.lambda_function
    content {
      id                  = lambda_function.value.id
      lambda_function_arn = lambda_function.value.lambda_function_arn
      events              = lambda_function.value.events
      filter_prefix       = lambda_function.value.filter_prefix
      filter_suffix       = lambda_function.value.filter_suffix
    }
  }

  dynamic "queue" {
    for_each = data.aws_s3_bucket_notification.existing.queue
    content {
      id            = queue.value.id
      queue_arn     = queue.value.queue_arn
      events        = queue.value.events
      filter_prefix = queue.value.filter_prefix
      filter_suffix = queue.value.filter_suffix
    }
  }

  dynamic "topic" {
    for_each = data.aws_s3_bucket_notification.existing.topic
    content {
      id            = topic.value.id
      topic_arn     = topic.value.topic_arn
      events        = topic.value.events
      filter_prefix = topic.value.filter_prefix
      filter_suffix = topic.value.filter_suffix
    }
  }
}
```

To add a new rule alongside existing ones, exclude IDs your resource owns from the iteration to avoid duplicates, and declare those rules separately:

```terraform
resource "aws_s3_bucket_notification" "example" {
  bucket = aws_s3_bucket.example.id

  dynamic "lambda_function" {
    for_each = [
      for f in data.aws_s3_bucket_notification.existing.lambda_function : f
      if f.id != "my-team-rule"
    ]
    content {
      id                  = lambda_function.value.id
      lambda_function_arn = lambda_function.value.lambda_function_arn
      events              = lambda_function.value.events
      filter_prefix       = lambda_function.value.filter_prefix
      filter_suffix       = lambda_function.value.filter_suffix
    }
  }

  lambda_function {
    id                  = "my-team-rule"
    lambda_function_arn = aws_lambda_function.mine.arn
    events              = ["s3:ObjectRemoved:*"]
  }
}
```

~> **Note:** The S3 API has no per-rule mutation primitive and no compare-and-swap, so two `terraform apply` runs from different state files writing to the same bucket can still race. For independent consumers of one bucket, [EventBridge](../r/s3_bucket_notification.html.markdown#emit-events-to-eventbridge) is generally a better fit.

## Argument Reference

This data source supports the following arguments:

* `bucket` - (Required) Name of the bucket.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `eventbridge` - Whether Amazon EventBridge notifications are enabled on this bucket.
* `lambda_function` - List of Lambda function notification configurations. See [`lambda_function`](#lambda_function) below.
* `queue` - List of SQS queue notification configurations. See [`queue`](#queue) below.
* `topic` - List of SNS topic notification configurations. See [`topic`](#topic) below.

### `lambda_function`

* `events` - [Events](https://docs.aws.amazon.com/AmazonS3/latest/userguide/notification-how-to-event-types-and-destinations.html) for which Amazon S3 sends notifications.
* `filter_prefix` - Object key name prefix.
* `filter_suffix` - Object key name suffix.
* `id` - Unique identifier for the notification configuration.
* `lambda_function_arn` - ARN of the Lambda function.

### `queue`

* `events` - [Events](https://docs.aws.amazon.com/AmazonS3/latest/userguide/notification-how-to-event-types-and-destinations.html) for which Amazon S3 sends notifications.
* `filter_prefix` - Object key name prefix.
* `filter_suffix` - Object key name suffix.
* `id` - Unique identifier for the notification configuration.
* `queue_arn` - ARN of the SQS queue.

### `topic`

* `events` - [Events](https://docs.aws.amazon.com/AmazonS3/latest/userguide/notification-how-to-event-types-and-destinations.html) for which Amazon S3 sends notifications.
* `filter_prefix` - Object key name prefix.
* `filter_suffix` - Object key name suffix.
* `id` - Unique identifier for the notification configuration.
* `topic_arn` - ARN of the SNS topic.
