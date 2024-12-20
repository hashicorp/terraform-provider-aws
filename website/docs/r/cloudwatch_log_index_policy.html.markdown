---
subcategory: "CloudWatch Logs"
layout: "aws"
page_title: "AWS: aws_cloudwatch_log_index_policy"
description: |-
  Terraform resource for managing an AWS CloudWatch Logs Index Policy.
---

# Resource: aws_cloudwatch_log_index_policy

Terraform resource for managing an AWS CloudWatch Logs Index Policy.

## Example Usage

### Basic Usage

```terraform
resource "aws_cloudwatch_log_group" "example" {
  name = "example"
}

resource "aws_cloudwatch_log_index_policy" "example" {
  log_group_name = aws_cloudwatch_log_group.example.name
  policy_document = jsonencode({
    Fields = ["eventName"]
  })
}
```

## Argument Reference

The following arguments are required:

* `log_group_name` - (Required) Log group name to set the policy for.
* `policy_document` - (Required) JSON policy document. This is a JSON formatted string.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CloudWatch Logs Index Policy using the `log_group_name`. For example:

```terraform
import {
  to = aws_cloudwatch_log_index_policy.example
  id = "/aws/log/group/name"
}
```

Using `terraform import`, import CloudWatch Logs Index Policy using the `log_group_name`. For example:

```console
% terraform import aws_cloudwatch_log_index_policy.example /aws/log/group/name
```
