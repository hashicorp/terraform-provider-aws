---
subcategory: "API Gateway"
layout: "aws"
page_title: "AWS: aws_api_gateway_account"
description: |-
  Provides a settings of an API Gateway Account.
---

# Resource: aws_api_gateway_account

Provides a settings of an API Gateway Account. Settings is applied region-wide per `provider` block.

-> **Note:** By default, destroying this resource will keep your account settings intact. Set `reset_on_delete` to `true` to reset the account setttings to default. In a future major version of the provider, destroying the resource will reset account settings.

## Example Usage

```terraform
resource "aws_api_gateway_account" "demo" {
  cloudwatch_role_arn = aws_iam_role.cloudwatch.arn
}

data "aws_iam_policy_document" "assume_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["apigateway.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
}

resource "aws_iam_role" "cloudwatch" {
  name               = "api_gateway_cloudwatch_global"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

data "aws_iam_policy_document" "cloudwatch" {
  statement {
    effect = "Allow"

    actions = [
      "logs:CreateLogGroup",
      "logs:CreateLogStream",
      "logs:DescribeLogGroups",
      "logs:DescribeLogStreams",
      "logs:PutLogEvents",
      "logs:GetLogEvents",
      "logs:FilterLogEvents",
    ]

    resources = ["*"]
  }
}
resource "aws_iam_role_policy" "cloudwatch" {
  name   = "default"
  role   = aws_iam_role.cloudwatch.id
  policy = data.aws_iam_policy_document.cloudwatch.json
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `cloudwatch_role_arn` - (Optional) ARN of an IAM role for CloudWatch (to allow logging & monitoring). See more [in AWS Docs](https://docs.aws.amazon.com/apigateway/latest/developerguide/how-to-stage-settings.html#how-to-stage-settings-console). Logging & monitoring can be enabled/disabled and otherwise tuned on the API Gateway Stage level.
* `reset_on_delete` - (Optional) If `true`, destroying the resource will reset account settings to default, otherwise account settings are not modified.
  Defaults to `false`.
  Will be removed in a future major version of the provider.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `api_key_version` - The version of the API keys used for the account.
* `throttle_settings` - Account-Level throttle settings. See exported fields below.
* `features` - A list of features supported for the account.

`throttle_settings` block exports the following:

* `burst_limit` - Absolute maximum number of times API Gateway allows the API to be called per second (RPS).
* `rate_limit` - Number of times API Gateway allows the API to be called per second on average (RPS).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import API Gateway Accounts using the account ID. For example:

```terraform
import {
  to = aws_api_gateway_account.demo
  id = "123456789012"
}
```

Using `terraform import`, import API Gateway Accounts using the account ID. For example:

```console
% terraform import aws_api_gateway_account.demo 123456789012
```
