---
subcategory: "CloudWatch Logs"
layout: "aws"
page_title: "AWS: aws_cloudwatch_log_destination_policy"
description: |-
  Provides a CloudWatch Logs destination policy.
---

# Resource: aws_cloudwatch_log_destination_policy

Provides a CloudWatch Logs destination policy resource.

## Example Usage

```terraform
resource "aws_cloudwatch_log_destination" "test_destination" {
  name       = "test_destination"
  role_arn   = aws_iam_role.iam_for_cloudwatch.arn
  target_arn = aws_kinesis_stream.kinesis_for_cloudwatch.arn
}

data "aws_iam_policy_document" "test_destination_policy" {
  statement {
    effect = "Allow"

    principals {
      type = "AWS"

      identifiers = [
        "123456789012",
      ]
    }

    actions = [
      "logs:PutSubscriptionFilter",
    ]

    resources = [
      aws_cloudwatch_log_destination.test_destination.arn,
    ]
  }
}

resource "aws_cloudwatch_log_destination_policy" "test_destination_policy" {
  destination_name = aws_cloudwatch_log_destination.test_destination.name
  access_policy    = data.aws_iam_policy_document.test_destination_policy.json
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `destination_name` - (Required) A name for the subscription filter
* `access_policy` - (Required) The policy document. This is a JSON formatted string.
* `force_update` - (Optional) Specify true if you are updating an existing destination policy to grant permission to an organization ID instead of granting permission to individual AWS accounts.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CloudWatch Logs destination policies using the `destination_name`. For example:

```terraform
import {
  to = aws_cloudwatch_log_destination_policy.test_destination_policy
  id = "test_destination"
}
```

Using `terraform import`, import CloudWatch Logs destination policies using the `destination_name`. For example:

```console
% terraform import aws_cloudwatch_log_destination_policy.test_destination_policy test_destination
```
