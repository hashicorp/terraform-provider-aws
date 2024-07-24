---
subcategory: "Shield"
layout: "aws"
page_title: "AWS: aws_shield_subscription"
description: |-
  Terraform resource for managing an AWS Shield Subscription.
---

# Resource: aws_shield_subscription

Terraform resource for managing an AWS Shield Subscription.

~> This resource creates a subscription to AWS Shield Advanced, which requires a 1 year subscription commitment with a monthly fee. Refer to the [AWS Shield Pricing](https://aws.amazon.com/shield/pricing/) page for more details.

~> Destruction of this resource will set `auto_renew` to `DISABLED`. Automatic renewal can only be disabled during the last 30 days of a subscription. To unsubscribe outside of this window, you must contact AWS Support. Set `skip_destroy` to `true` to skip modifying the `auto_renew` argument during destruction.

## Example Usage

### Basic Usage

```terraform
resource "aws_shield_subscription" "example" {
  auto_renew = "ENABLED"
}
```

## Argument Reference

The following arguments are optional:

* `auto_renew` - (Optional) Toggle for automated renewal of the subscription. Valid values are `ENABLED` or `DISABLED`. Default is `ENABLED`.
* `skip_destroy` - (Optional) Skip attempting to disable automated renewal upon destruction. If set to `true`, the `auto_renew` value will be left as-is and the resource will simply be removed from state.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - AWS Account ID.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Shield Subscription using the `id`. For example:

```terraform
import {
  to = aws_shield_subscription.example
  id = "012345678901"
}
```

Using `terraform import`, import Shield Subscription using the `id`. For example:

```console
% terraform import aws_shield_subscription.example 012345678901
```
