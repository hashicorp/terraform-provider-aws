---
subcategory: "Directory Service"
layout: "aws"
page_title: "AWS: aws_directory_service_log_subscription"
description: |-
  Provides a Log subscription for AWS Directory Service that pushes logs to cloudwatch.
---

# Resource: aws_directory_service_log_subscription

Provides a Log subscription for AWS Directory Service that pushes logs to cloudwatch.

## Example Usage

```terraform
resource "aws_cloudwatch_log_group" "example" {
  name              = "/aws/directoryservice/${aws_directory_service_directory.example.id}"
  retention_in_days = 14
}

data "aws_iam_policy_document" "ad-log-policy" {
  statement {
    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents",
    ]

    principals {
      identifiers = ["ds.amazonaws.com"]
      type        = "Service"
    }

    resources = ["${aws_cloudwatch_log_group.example.arn}:*"]

    effect = "Allow"
  }
}

resource "aws_cloudwatch_log_resource_policy" "ad-log-policy" {
  policy_document = data.aws_iam_policy_document.ad-log-policy.json
  policy_name     = "ad-log-policy"
}

resource "aws_directory_service_log_subscription" "example" {
  directory_id   = aws_directory_service_directory.example.id
  log_group_name = aws_cloudwatch_log_group.example.name
}
```

## Argument Reference

This resource supports the following arguments:

* `directory_id` - (Required) ID of directory.
* `log_group_name` - (Required) Name of the cloudwatch log group to which the logs should be published. The log group should be already created and the directory service principal should be provided with required permission to create stream and publish logs. Changing this value would delete the current subscription and create a new one. A directory can only have one log subscription at a time.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Directory Service Log Subscriptions using the directory id. For example:

```terraform
import {
  to = aws_directory_service_log_subscription.msad
  id = "d-1234567890"
}
```

Using `terraform import`, import Directory Service Log Subscriptions using the directory id. For example:

```console
% terraform import aws_directory_service_log_subscription.msad d-1234567890
```
