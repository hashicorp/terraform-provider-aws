---
subcategory: "CloudWatch Logs"
layout: "aws"
page_title: "AWS: aws_cloudwatch_log_resource_policy"
description: |-
  Provides a resource to manage a CloudWatch log resource policy
---

# Resource: aws_cloudwatch_log_resource_policy

Provides a resource to manage a CloudWatch log resource policy.

## Example Usage

### Elasticsearch Log Publishing

```terraform
data "aws_iam_policy_document" "elasticsearch-log-publishing-policy" {
  statement {
    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents",
      "logs:PutLogEventsBatch",
    ]

    resources = ["arn:aws:logs:*"]

    principals {
      identifiers = ["es.amazonaws.com"]
      type        = "Service"
    }
  }
}

resource "aws_cloudwatch_log_resource_policy" "elasticsearch-log-publishing-policy" {
  policy_document = data.aws_iam_policy_document.elasticsearch-log-publishing-policy.json
  policy_name     = "elasticsearch-log-publishing-policy"
}
```

### Route53 Query Logging

```terraform
data "aws_iam_policy_document" "route53-query-logging-policy" {
  statement {
    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents",
    ]

    resources = ["arn:aws:logs:*:*:log-group:/aws/route53/*"]

    principals {
      identifiers = ["route53.amazonaws.com"]
      type        = "Service"
    }
  }
}

resource "aws_cloudwatch_log_resource_policy" "route53-query-logging-policy" {
  policy_document = data.aws_iam_policy_document.route53-query-logging-policy.json
  policy_name     = "route53-query-logging-policy"
}
```

## Argument Reference

This resource supports the following arguments:

* `policy_document` - (Required) Details of the resource policy, including the identity of the principal that is enabled to put logs to this account. This is formatted as a JSON string. Maximum length of 5120 characters.
* `policy_name` - (Required) Name of the resource policy.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The name of the CloudWatch log resource policy

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CloudWatch log resource policies using the policy name. For example:

```terraform
import {
  to = aws_cloudwatch_log_resource_policy.MyPolicy
  id = "MyPolicy"
}
```

Using `terraform import`, import CloudWatch log resource policies using the policy name. For example:

```console
% terraform import aws_cloudwatch_log_resource_policy.MyPolicy MyPolicy
```
