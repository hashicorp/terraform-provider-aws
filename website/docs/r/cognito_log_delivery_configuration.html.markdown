---
subcategory: "Cognito IDP (Identity Provider)"
layout: "aws"
page_title: "AWS: aws_cognito_log_delivery_configuration"
description: |-
  Manages an AWS Cognito IDP (Identity Provider) Log Delivery Configuration.
---

# Resource: aws_cognito_log_delivery_configuration

Manages an AWS Cognito IDP (Identity Provider) Log Delivery Configuration.

## Example Usage

### Basic Usage with CloudWatch Logs

```terraform
resource "aws_cognito_user_pool" "example" {
  name = "example"
}

resource "aws_cloudwatch_log_group" "example" {
  name = "example"
}

resource "aws_cognito_log_delivery_configuration" "example" {
  user_pool_id = aws_cognito_user_pool.example.id

  log_configurations {
    event_source = "userNotification"
    log_level    = "ERROR"

    cloud_watch_logs_configuration {
      log_group_arn = aws_cloudwatch_log_group.example.arn
    }
  }
}
```

### Multiple Log Configurations with Different Destinations

```terraform
resource "aws_cognito_user_pool" "example" {
  name = "example"
}

resource "aws_cloudwatch_log_group" "example" {
  name = "example"
}

resource "aws_s3_bucket" "example" {
  bucket        = "example-bucket"
  force_destroy = true
}

resource "aws_iam_role" "firehose" {
  name = "firehose-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "firehose.amazonaws.com"
        }
      }
    ]
  })
}

resource "aws_iam_role_policy" "firehose" {
  name = "firehose-policy"
  role = aws_iam_role.firehose.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3:AbortMultipartUpload",
          "s3:GetBucketLocation",
          "s3:GetObject",
          "s3:ListBucket",
          "s3:ListBucketMultipartUploads",
          "s3:PutObject"
        ]
        Resource = [
          aws_s3_bucket.example.arn,
          "${aws_s3_bucket.example.arn}/*"
        ]
      }
    ]
  })
}

resource "aws_kinesis_firehose_delivery_stream" "example" {
  name        = "example-stream"
  destination = "extended_s3"

  extended_s3_configuration {
    role_arn   = aws_iam_role.firehose.arn
    bucket_arn = aws_s3_bucket.example.arn
  }
}

resource "aws_cognito_log_delivery_configuration" "example" {
  user_pool_id = aws_cognito_user_pool.example.id

  log_configurations {
    event_source = "userNotification"
    log_level    = "INFO"

    cloud_watch_logs_configuration {
      log_group_arn = aws_cloudwatch_log_group.example.arn
    }
  }

  log_configurations {
    event_source = "userAuthEvents"
    log_level    = "ERROR"

    firehose_configuration {
      stream_arn = aws_kinesis_firehose_delivery_stream.example.arn
    }
  }
}
```

### S3 Configuration

```terraform
resource "aws_cognito_user_pool" "example" {
  name = "example"
}

resource "aws_s3_bucket" "example" {
  bucket        = "example-bucket"
  force_destroy = true
}

resource "aws_cognito_log_delivery_configuration" "example" {
  user_pool_id = aws_cognito_user_pool.example.id

  log_configurations {
    event_source = "userNotification"
    log_level    = "ERROR"

    s3_configuration {
      bucket_arn = aws_s3_bucket.example.arn
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `user_pool_id` - (Required) The ID of the user pool for which to configure log delivery.

The following arguments are optional:

* `log_configurations` - (Optional) Configuration block for log delivery. At least one configuration block is required. See [Log Configurations](#log-configurations) below.
* `region` - (Optional) The AWS region.

### Log Configurations

The `log_configurations` block supports the following:

* `event_source` - (Required) The event source to configure logging for. Valid values are `userNotification` and `userAuthEvents`.
* `log_level` - (Required) The log level to set for the event source. Valid values are `ERROR` and `INFO`.
* `cloud_watch_logs_configuration` - (Optional) Configuration for CloudWatch Logs delivery. See [CloudWatch Logs Configuration](#cloudwatch-logs-configuration) below.
* `firehose_configuration` - (Optional) Configuration for Kinesis Data Firehose delivery. See [Firehose Configuration](#firehose-configuration) below.
* `s3_configuration` - (Optional) Configuration for S3 delivery. See [S3 Configuration](#s3-configuration) below.

~> **Note:** At least one destination configuration (`cloud_watch_logs_configuration`, `firehose_configuration`, or `s3_configuration`) must be specified for each log configuration.

#### CloudWatch Logs Configuration

The `cloud_watch_logs_configuration` block supports the following:

* `log_group_arn` - (Optional) The ARN of the CloudWatch Logs log group to which the logs should be delivered.

#### Firehose Configuration

The `firehose_configuration` block supports the following:

* `stream_arn` - (Optional) The ARN of the Kinesis Data Firehose delivery stream to which the logs should be delivered.

#### S3 Configuration

The `s3_configuration` block supports the following:

* `bucket_arn` - (Optional) The ARN of the S3 bucket to which the logs should be delivered.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_cognito_log_delivery_configuration.example
  identity = {
    user_pool_id = "us-west-2_example123"
  }
}

resource "aws_cognito_log_delivery_configuration" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `user_pool_id` (String) ID of the Cognito User Pool.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Cognito IDP (Identity Provider) Log Delivery Configuration using the `user_pool_id`. For example:

```terraform
import {
  to = aws_cognito_log_delivery_configuration.example
  id = "us-west-2_example123"
}
```

Using `terraform import`, import Cognito IDP (Identity Provider) Log Delivery Configuration using the `user_pool_id`. For example:

```console
% terraform import aws_cognito_log_delivery_configuration.example us-west-2_example123
```
