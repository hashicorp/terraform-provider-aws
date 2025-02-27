---
subcategory: "Bedrock"
layout: "aws"
page_title: "AWS: aws_bedrock_model_invocation_logging_configuration"
description: |-
  Manages Bedrock model invocation logging configuration.
---

# Resource: aws_bedrock_model_invocation_logging_configuration

Manages Bedrock model invocation logging configuration.

~> Model invocation logging is configured per AWS region. To avoid overwriting settings, this resource should not be defined in multiple configurations.

## Example Usage

### Basic Usage

```terraform
data "aws_caller_identity" "current" {}

resource aws_s3_bucket example {
  bucket        = "example"
  force_destroy = true
  lifecycle {
    ignore_changes = [
      tags["CreatorId"], tags["CreatorName"],
    ]
  }
}

resource "aws_s3_bucket_policy" "example" {
  bucket = aws_s3_bucket.example.bucket

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "bedrock.amazonaws.com"
      },
      "Action": [
        "s3:*"
      ],
      "Resource": [
        "${aws_s3_bucket.example.arn}/*"
      ],
      "Condition": {
        "StringEquals": {
          "aws:SourceAccount": "${data.aws_caller_identity.current.account_id}"
        },
        "ArnLike": {
          "aws:SourceArn": "arn:aws:bedrock:us-east-1:${data.aws_caller_identity.current.account_id}:*"
        }
      }
    }
  ]
}
EOF
}

resource "aws_bedrock_model_invocation_logging_configuration" "example" {
  depends_on = [
    aws_s3_bucket_policy.example
  ]

  logging_config {
    embedding_data_delivery_enabled = true
    image_data_delivery_enabled     = true
    text_data_delivery_enabled      = true
    video_data_delivery_enabled     = true
    s3_config {
      bucket_name = aws_s3_bucket.example.id
      key_prefix  = "bedrock"
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `logging_config` - (Required) The logging configuration values to set. See [`logging_config` Block](#logging_config-block) for details.

### `logging_config` Block

The `logging_config` configuration block supports the following arguments:

* `cloudwatch_config` – (Optional) CloudWatch logging configuration. See [`cloudwatch_config` Block](#cloudwatch_config-block) for details.
* `embedding_data_delivery_enabled` – (Optional) Set to include embeddings data in the log delivery. Defaults to `true`.
* `image_data_delivery_enabled` – (Optional) Set to include image data in the log delivery. Defaults to `true`.
* `s3_config` – (Optional) S3 configuration for storing log data. See [`s3_config` Block](#s3_config-block) for details.
* `text_data_delivery_enabled` – (Optional) Set to include text data in the log delivery. Defaults to `true`.
* `video_data_delivery_enabled` – (Optional) Set to include text data in the log delivery. Defaults to `true`.

### `cloudwatch_config` Block

The `cloudwatch_config` configuration block supports the following arguments:

* `large_data_delivery_s3_config` – (Optional) S3 configuration for delivering a large amount of data. See [`large_data_delivery_s3_config` Block](#large_data_delivery_s3_config-block) for details.
* `log_group_name` – (Required) Log group name.
* `role_arn` – (Optional) The role ARN.

### `large_data_delivery_s3_config` Block

The `large_data_delivery_s3_config` configuration block supports the following arguments:

* `bucket_name` – (Required) S3 bucket name.
* `key_prefix` – (Optional) S3 prefix.

### `s3_config` Block

The `s3_config` configuration block supports the following arguments:

* `bucket_name` – (Required) S3 bucket name.
* `key_prefix` – (Optional) S3 prefix.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - AWS Region in which logging is configured.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Bedrock Invocation Logging Configuration using the `id` set to the AWS Region. For example:

```terraform
import {
  to = aws_bedrock_model_invocation_logging_configuration.my_config
  id = "us-east-1"
}
```

Using `terraform import`, import Bedrock custom model using the `id` set to the AWS Region. For example:

```console
% terraform import aws_bedrock_model_invocation_logging_configuration.my_config us-east-1
```
