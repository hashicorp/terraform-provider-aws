---
subcategory: "Amazon Bedrock"
layout: "aws"
page_title: "AWS: aws_bedrock_model_invocation_logging_configuration"
description: |-
  Manages Bedrock model invocation logging configuration.
---

# Resource: aws_bedrock_model_invocation_logging_configuration

Manages Bedrock model invocation logging configuration.

~> Model invocation logging is configured per AWS region. To avoid overwriting settings, this resource should not be defined in multiple configurations.

## Example Usage

```terraform
data "aws_caller_identity" "current" {}

resource aws_s3_bucket bedrock_logging {
  bucket        = "bedrock-logging-%[1]s"
  force_destroy = true
  lifecycle {
    ignore_changes = [
      tags["CreatorId"], tags["CreatorName"],
    ]
  }
}

resource "aws_s3_bucket_policy" "bedrock_logging" {
  bucket = aws_s3_bucket.bedrock_logging.bucket

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
        "${aws_s3_bucket.bedrock_logging.arn}/*"
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

resource "aws_bedrock_model_invocation_logging_configuration" "test" {
  logging_config {
    embedding_data_delivery_enabled = true
    image_data_delivery_enabled     = true
    text_data_delivery_enabled      = true
    s3_config {
      bucket_name = aws_s3_bucket.bedrock_logging.id
      key_prefix  = "bedrock"
    }
  }
  depends_on = [
    aws_s3_bucket_policy.bedrock_logging
  ]
}
```

## Argument Reference

The following arguments are required:

* `logging_config` - The logging configuration values to set. See [`logging_config`](#logging_config-argument-reference).

### `logging_config` Argument Reference

The following arguments are optional:

* `cloudwatch_config` – CloudWatch logging configuration. See [`cloudwatch_config`](#cloudwatch_config-argument-reference).
* `embedding_data_delivery_enabled` – Set to include embeddings data in the log delivery.
* `image_data_delivery_enabled` – Set to include image data in the log delivery.
* `s3_config` – S3 configuration for storing log data. See [`s3_config`](#s3_config-argument-reference).
* `text_data_delivery_enabled` – Set to include text data in the log delivery.

### `cloudwatch_config` Argument Reference

The following arguments are required:

* `log_group_name` – Log group name.
* `role_arn` – IAM Role ARN.

The following arguments are optional:

* `large_data_delivery_s3_config` – S3 configuration for delivering a large amount of data. See [`s3_config`](#s3_config-argument-reference).

### `s3_config` Argument Reference

The following arguments are required:

* `bucket_name` – S3 bucket name.

The following arguments are optional:

* `key_prefix` – S3 object key prefix.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - AWS region in which logging is configured.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Bedrock Invocation Logging Configuration using the `id` set to the AWS region. For example:

```terraform
import {
  to = aws_bedrock_model_invocation_logging_configuration.my_config
  id = "us-east-1"
}
```

Using `terraform import`, import Bedrock custom model using the `id` set to the AWS region. For example:

```console
% terraform import aws_bedrock_model_invocation_logging_configuration.my_config us-east-1
```
