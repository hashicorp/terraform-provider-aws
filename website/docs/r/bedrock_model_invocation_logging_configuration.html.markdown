---
subcategory: "Amazon Bedrock"
layout: "aws"
page_title: "AWS: aws_bedrock_model_invocation_logging_configuration"
description: |-
  Manages Bedrock model invocation logging configuration.
---

# Resource: aws_bedrock_model_invocation_logging_configuration

Manages Bedrock model invocation logging configuration.

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

* `logging_config` - The logging configuration values to set. Type: LoggingConfig object.

The following arguments are optional:

NONE

### LoggingConfig Object

The following arguments are required:

NONE

The following arguments are optional:

* `cloud_watch_config` – CloudWatch logging configuration. Type: CloudWatchConfig object.
* `embedding_data_delivery_enabled` – Set to include embeddings data in the log delivery. Type: Boolean. Default: True
* `image_data_delivery_enabled` – Set to include image data in the log delivery. Type: Boolean. Default: True
* `textDataDeliveryEnabled` – Set to include text data in the log delivery. Type: Boolean. Default: True
* `s3_config` – S3 configuration for storing log data. Type: S3Config object.

### CloudWatchConfig Object

The following arguments are required:

* `log_group_name` – The log group name. Type: String. Length Constraints: Minimum length of 1. Maximum length of 512.
* `role_arn` – The role ARN.. Type: String. Length Constraints: Minimum length of 0. Maximum length of 2048. Pattern: ^arn:aws(-[^:]+)?:iam::([0-9]{12})?:role/.+$

The following arguments are optional:

* `large_data_delivery_s3_config` – S3 configuration for delivering a large amount of data. Type: S3Config.

### S3Config Object

The following arguments are required:

* `bucket_name` – S3 bucket name. Type: String. Length Constraints: Minimum length of 3. Maximum length of 63.

The following arguments are optional:

* `key_prefix` – S3 prefix. Type: String. Length Constraints: Minimum length of 0. Maximum length of 1024.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Bedrock Custom Model using the `model_id`:

```terraform
import {
  to = aws_bedrock_model_invocation_logging_configuration.my_config
  id = "us-east-1"
}
```

Using `terraform import`, import Bedrock custom model using the `model_id`:

```console
% terraform import aws_bedrock_model_invocation_logging_configuration.my_config us-east-1
```
