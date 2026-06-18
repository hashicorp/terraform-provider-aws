---
subcategory: "S3 (Simple Storage)"
layout: "aws"
page_title: "AWS: aws_s3_bucket_notification"
description: |-
  Provides the existing notification configuration of an S3 bucket.
---

# Data Source: aws_s3_bucket_notification

Provides the existing notification configuration of an S3 bucket.

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
