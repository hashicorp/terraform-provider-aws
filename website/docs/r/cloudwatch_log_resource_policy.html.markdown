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

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `policy_document` - (Required) Details of the resource policy, including the identity of the principal that is enabled to put logs to this account. This is formatted as a JSON string. Maximum length of 5120 characters.
* `policy_name` - (Optional) Name of the resource policy. Exactly one of `policy_name` or `resource_arn` must be specified. Note that the number of resource policies without `resource_arn` is limited to 10 per region.
* `resource_arn` - (Optional) ARN of the CloudWatch Logs resource to which the resource policy is attached. Exactly one of `policy_name` or `resource_arn` must be specified. Only one policy can be attached per log group resource ARN.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The name of the CloudWatch log resource policy when `resource_arn` is not specified, or the ARN of the CloudWatch log group when `resource_arn` is specified.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CloudWatch Logs resource policies using the policy name for policies that do not specify a CloudWatch Logs resource ARN, or the ARN of the CloudWatch Logs resource to which the policy is attached for policies that do. For example:

```terraform
import {
  to = aws_cloudwatch_log_resource_policy.my_policy_without_resource_arn
  id = "my_policy"
}
```

```terraform
import {
  to = aws_cloudwatch_log_resource_policy.my_policy_with_resource_arn
  id = "arn:aws:logs:us-west-2:123456789012:log-group:/my-log-group"
}
```

Using `terraform import`, import CloudWatch log resource policies using the policy name for policies that do not specify a CloudWatch Logs resource ARN, or the ARN of the CloudWatch Logs resource to which the policy is attached for policies that do. For example:

```console
% terraform import aws_cloudwatch_log_resource_policy.my_policy_without_resource_arn my_policy
```

```console
% terraform import aws_cloudwatch_log_resource_policy.my_policy_with_resource_arn "arn:aws:logs:us-west-2:123456789012:log-group:/my-log-group"
```
