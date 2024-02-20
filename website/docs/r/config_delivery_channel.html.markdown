---
subcategory: "Config"
layout: "aws"
page_title: "AWS: aws_config_delivery_channel"
description: |-
  Provides an AWS Config Delivery Channel.
---

# Resource: aws_config_delivery_channel

Provides an AWS Config Delivery Channel.

~> **Note:** Delivery Channel requires a [Configuration Recorder](/docs/providers/aws/r/config_configuration_recorder.html) to be present. Use of `depends_on` (as shown below) is recommended to avoid race conditions.

## Example Usage

```terraform
resource "aws_config_delivery_channel" "foo" {
  name           = "example"
  s3_bucket_name = aws_s3_bucket.b.bucket
  depends_on     = [aws_config_configuration_recorder.foo]
}

resource "aws_s3_bucket" "b" {
  bucket        = "example-awsconfig"
  force_destroy = true
}

resource "aws_config_configuration_recorder" "foo" {
  name     = "example"
  role_arn = aws_iam_role.r.arn
}

data "aws_iam_policy_document" "assume_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["config.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
}

resource "aws_iam_role" "r" {
  name               = "awsconfig-example"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

data "aws_iam_policy_document" "p" {
  statement {
    effect  = "Allow"
    actions = ["s3:*"]
    resources = [
      aws_s3_bucket.b.arn,
      "${aws_s3_bucket.b.arn}/*"
    ]
  }
}

resource "aws_iam_role_policy" "p" {
  name   = "awsconfig-example"
  role   = aws_iam_role.r.id
  policy = data.aws_iam_policy_document.p.json
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Optional) The name of the delivery channel. Defaults to `default`. Changing it recreates the resource.
* `s3_bucket_name` - (Required) The name of the S3 bucket used to store the configuration history.
* `s3_key_prefix` - (Optional) The prefix for the specified S3 bucket.
* `s3_kms_key_arn` - (Optional) The ARN of the AWS KMS key used to encrypt objects delivered by AWS Config. Must belong to the same Region as the destination S3 bucket.
* `sns_topic_arn` - (Optional) The ARN of the SNS topic that AWS Config delivers notifications to.
* `snapshot_delivery_properties` - (Optional) Options for how AWS Config delivers configuration snapshots. See below

### `snapshot_delivery_properties`

* `delivery_frequency` - (Optional) - The frequency with which AWS Config recurringly delivers configuration snapshotsE.g., `One_Hour` or `Three_Hours`. Valid values are listed [here](https://docs.aws.amazon.com/config/latest/APIReference/API_ConfigSnapshotDeliveryProperties.html#API_ConfigSnapshotDeliveryProperties_Contents).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The name of the delivery channel.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Delivery Channel using the name. For example:

```terraform
import {
  to = aws_config_delivery_channel.foo
  id = "example"
}
```

Using `terraform import`, import Delivery Channel using the name. For example:

```console
% terraform import aws_config_delivery_channel.foo example
```
