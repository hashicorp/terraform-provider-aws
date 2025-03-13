---
subcategory: "CloudWatch Logs"
layout: "aws"
page_title: "AWS: aws_cloudwatch_log_delivery_destination_policy"
description: |-
  Terraform resource for managing an AWS CloudWatch Logs Delivery Destination Policy.
---

# Resource: aws_cloudwatch_log_delivery_destination_policy

Terraform resource for managing an AWS CloudWatch Logs Delivery Destination Policy.

## Example Usage

### Basic Usage

```terraform
resource "aws_cloudwatch_log_delivery_destination_policy" "example" {
  delivery_destination_name   = aws_cloudwatch_log_delivery_destination.example.name
  delivery_destination_policy = data.aws_iam_policy_document.example.json
}
```

## Argument Reference

This resource supports the following arguments:

* `delivery_destination_name` - (Required) The name of the delivery destination to assign this policy to.
* `delivery_destination_policy` - (Required) The contents of the policy.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CloudWatch Logs Delivery Destination Policy using the `delivery_destination_name`. For example:

```terraform
import {
  to = aws_cloudwatch_log_delivery_destination_policy.example
  id = "example"
}
```

Using `terraform import`, import CloudWatch Logs Delivery Destination Policy using the `delivery_destination_name`. For example:

```console
% terraform import aws_cloudwatch_log_delivery_destination_policy.example example
```
