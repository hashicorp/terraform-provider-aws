---
subcategory: "S3 (Simple Storage)"
layout: "aws"
page_title: "AWS: aws_s3_bucket_notification"
description: |-
  Provides details about a S3 Bucket Notification Configuration
---

# Data Source: aws_s3_bucket_notification

Provides details about a S3 Bucket Notification Configuration

This resource may prove useful when you want to have multiple S3 notifications in different repo's.

## Example Usage

```terraform
data "aws_s3_bucket_notification" "test" {
  bucket = "test"
}
```

## Argument Reference

The following argument is required:
* `bucket` - (Required) Name of the bucket for notification configuration.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

### `lambda_function_configurations`

* `events` - [Event](http://docs.aws.amazon.com/AmazonS3/latest/dev/NotificationHowTo.html#notification-how-to-event-types-and-destinations) for which to send notifications.
* `filter_prefix` - Object key name prefix.
* `filter_suffix` - Object key name suffix.
* `id` - Unique identifier for each of the notification configurations.
* `lambda_function_arn` - Lambda function ARN.

### `queue_configurations`

* `events` - Specifies [event](http://docs.aws.amazon.com/AmazonS3/latest/dev/NotificationHowTo.html#notification-how-to-event-types-and-destinations) for which to send notifications.
* `filter_prefix` - Object key name prefix.
* `filter_suffix` - Object key name suffix.
* `id` - Unique identifier for each of the notification configurations.
* `queue_arn` - SQS queue ARN.

### `topic_configurations`

* `events` - [Event](http://docs.aws.amazon.com/AmazonS3/latest/dev/NotificationHowTo.html#notification-how-to-event-types-and-destinations) for which to send notifications.
* `filter_prefix` - Object key name prefix.
* `filter_suffix` - Object key name suffix.
* `id` - Unique identifier for each of the notification configurations.
* `topic_arn` - SNS topic ARN.