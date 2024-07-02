---
subcategory: "Shield"
layout: "aws"
page_title: "AWS: aws_shield_subscription"
description: |-
  Terraform resource for managing an AWS Shield Subscription.
---

# Resource: aws_shield_subscription

Terraform resource for managing an AWS Shield Subscription.

## Example Usage

### Basic Usage

```terraform
resource "aws_shield_subscription" "example" {
  auto_renew = "ENABLED"
}
```

## Argument Reference

The following arguments are optional:

* `auto_renew` - (Optional) automated renewal for the subscription. Valid values are `ENABLED` or `DISABLED`. Default is `ENABLED`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Account ID is used.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Shield Subscription using the `AWS Account ID`. For example:

```terraform
import {
  to = aws_shield_subscription.example
  id = "1234567890"
}
```

Using `terraform import`, import Shield Subscription using the `AWS Account ID`. For example:

```console
% terraform import aws_shield_subscription.example 1234567890
```
