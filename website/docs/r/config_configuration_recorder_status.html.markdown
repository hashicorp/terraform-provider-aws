---
subcategory: "Config"
layout: "aws"
page_title: "AWS: aws_config_configuration_recorder_status"
description: |-
  Manages status of an AWS Config Configuration Recorder.
---

# Resource: aws_config_configuration_recorder_status

Manages status (recording / stopped) of an AWS Config Configuration Recorder.

~> **Note:** Starting Configuration Recorder requires a [Delivery Channel](/docs/providers/aws/r/config_delivery_channel.html) to be present. Use of `depends_on` (as shown below) is recommended to avoid race conditions.

## Example Usage

```terraform
resource "aws_config_configuration_recorder_status" "foo" {
  name       = aws_config_configuration_recorder.foo.name
  is_enabled = true
  depends_on = [aws_config_delivery_channel.foo]
}

resource "aws_iam_role_policy_attachment" "a" {
  role       = aws_iam_role.r.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWS_ConfigRole"
}

resource "aws_s3_bucket" "b" {
  bucket = "awsconfig-example"
}

resource "aws_config_delivery_channel" "foo" {
  name           = "example"
  s3_bucket_name = aws_s3_bucket.b.bucket
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
  name               = "example-awsconfig"
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

* `name` - (Required) The name of the recorder
* `is_enabled` - (Required) Whether the configuration recorder should be enabled or disabled.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Configuration Recorder Status using the name of the Configuration Recorder. For example:

```terraform
import {
  to = aws_config_configuration_recorder_status.foo
  id = "example"
}
```

Using `terraform import`, import Configuration Recorder Status using the name of the Configuration Recorder. For example:

```console
% terraform import aws_config_configuration_recorder_status.foo example
```
