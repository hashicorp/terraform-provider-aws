---
subcategory: "API Gateway"
layout: "aws"
page_title: "AWS: aws_api_gateway_account"
description: |-
  Provides a settings of an API Gateway Account.
---

# Resource: aws_api_gateway_account

Provides a settings of an API Gateway Account. Settings is applied region-wide per `provider` block.

-> **Note:** As there is no API method for deleting account settings or resetting it to defaults, destroying this resource will keep your account settings intact

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

* `cloudwatch_role_arn` - (Optional) ARN of an IAM role for CloudWatch (to allow logging & monitoring). See more [in AWS Docs](https://docs.aws.amazon.com/apigateway/latest/developerguide/how-to-stage-settings.html#how-to-stage-settings-console). Logging & monitoring can be enabled/disabled and otherwise tuned on the API Gateway Stage level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `throttle_settings` - Account-Level throttle settings. See exported fields below.

`throttle_settings` block exports the following:

* `burst_limit` - Absolute maximum number of times API Gateway allows the API to be called per second (RPS).
* `rate_limit` - Number of times API Gateway allows the API to be called per second on average (RPS).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import API Gateway Accounts using the word `api-gateway-account`. For example:

```terraform
import {
  to = aws_api_gateway_account.demo
  id = "api-gateway-account"
}
```

Using `terraform import`, import API Gateway Accounts using the word `api-gateway-account`. For example:

```console
% terraform import aws_api_gateway_account.demo api-gateway-account
```
