---
subcategory: "API Gateway"
layout: "aws"
page_title: "AWS: aws_api_gateway_rest_api_policy"
description: |-
  Provides an API Gateway REST API Policy.
---

# Resource: aws_api_gateway_rest_api_policy

Provides an API Gateway REST API Policy.

-> **Note:** Amazon API Gateway Version 1 resources are used for creating and deploying REST APIs. To create and deploy WebSocket and HTTP APIs, use Amazon API Gateway Version 2 [resources](/docs/providers/aws/r/apigatewayv2_api.html).

## Example Usage

### Basic

```terraform
resource "aws_api_gateway_rest_api" "test" {
  name = "example-rest-api"
}

data "aws_iam_policy_document" "test" {
  statement {
    effect = "Allow"

    principals {
      type        = "AWS"
      identifiers = ["*"]
    }

    actions   = ["execute-api:Invoke"]
    resources = [aws_api_gateway_rest_api.test.execution_arn]

    condition {
      test     = "IpAddress"
      variable = "aws:SourceIp"
      values   = ["123.123.123.123/32"]
    }
  }
}
resource "aws_api_gateway_rest_api_policy" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  policy      = data.aws_iam_policy_document.test.json
}
```

## Argument Reference

The following arguments are supported:

* `rest_api_id` - (Required) ID of the REST API.
* `policy` - (Required) JSON formatted policy document that controls access to the API Gateway. For more information about building AWS IAM policy documents with Terraform, see the [AWS IAM Policy Document Guide](https://learn.hashicorp.com/terraform/aws/iam-policy)

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the REST API

## Import

`aws_api_gateway_rest_api_policy` can be imported by using the REST API ID, e.g.,

```
$ terraform import aws_api_gateway_rest_api_policy.example 12345abcde
```
