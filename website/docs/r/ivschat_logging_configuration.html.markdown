---
subcategory: "IVS (Interactive Video) Chat"
layout: "aws"
page_title: "AWS: aws_ivschat_logging_configuration"
description: |-
  Terraform resource for managing an AWS IVS (Interactive Video) Chat Logging Configuration.
---

# Resource: aws_ivschat_logging_configuration

Terraform resource for managing an AWS IVS (Interactive Video) Chat Logging Configuration.

## Example Usage

### Basic Usage - Logging to CloudWatch

```terraform
resource "aws_cloudwatch_log_group" "example" {}

resource "aws_ivschat_logging_configuration" "example" {
  destination_configuration {
    cloudwatch_logs {
      log_group_name = aws_cloudwatch_log_group.example.name
    }
  }
}
```

### Basic Usage - Logging to Kinesis Firehose with Extended S3

```terraform
resource "aws_kinesis_firehose_delivery_stream" "example" {
  name        = "terraform-kinesis-firehose-extended-s3-example-stream"
  destination = "extended_s3"

  extended_s3_configuration {
    role_arn   = aws_iam_role.example.arn
    bucket_arn = aws_s3_bucket.example.arn
  }

  tags = {
    "LogDeliveryEnabled" = "true"
  }
}

resource "aws_s3_bucket" "example" {
  bucket_prefix = "tf-ivschat-logging-bucket"
}

resource "aws_s3_bucket_acl" "example" {
  bucket = aws_s3_bucket.example.id
  acl    = "private"
}

resource "aws_iam_role" "example" {
  name = "firehose_example_role"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "firehose.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_ivschat_logging_configuration" "example" {
  destination_configuration {
    firehose {
      delivery_stream_name = aws_kinesis_firehose_delivery_stream.example.name
    }
  }
}
```

### Basic Usage - Logging to S3

```terraform
resource "aws_s3_bucket" "example" {
  bucket_name   = "tf-ivschat-logging"
  force_destroy = true
}

resource "aws_ivschat_logging_configuration" "example" {
  destination_configuration {
    s3 {
      bucket_name = aws_s3_bucket.example.id
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `destination_configuration` - (Required) Object containing destination configuration for where chat activity will be logged. This object must contain exactly one of the following children arguments:
    * `cloudwatch_logs` - An Amazon CloudWatch Logs destination configuration where chat activity will be logged.
        * `log_group_name` - Name of the Amazon Cloudwatch Logs destination where chat activity will be logged.
    * `firehose` - An Amazon Kinesis Data Firehose destination configuration where chat activity will be logged.
        * `delivery_stream_name` - Name of the Amazon Kinesis Firehose delivery stream where chat activity will be logged.
    * `s3` - An Amazon S3 destination configuration where chat activity will be logged.
        * `bucket_name` - Name of the Amazon S3 bucket where chat activity will be logged.

The following arguments are optional:

* `name` - (Optional) Logging Configuration name.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the Logging Configuration.
* `id` - ID of the Logging Configuration.
* `state` - State of the Logging Configuration.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts):

* `create` - (Default `5m`)
* `update` - (Default `5m`)
* `delete` - (Default `5m`)

## Import

IVS (Interactive Video) Chat Logging Configuration can be imported using the ARN, e.g.,

```
$ terraform import aws_ivschat_logging_configuration.example arn:aws:ivschat:us-west-2:326937407773:logging-configuration/MMUQc8wcqZmC
```
