---
subcategory: "Pricing Plan Manager"
layout: "aws"
page_title: "AWS: aws_pricingplanmanager_subscription"
description: |-
  Lists Pricing Plan Manager Subscription resources.
---

# List Resource: aws_pricingplanmanager_subscription

Lists Pricing Plan Manager Subscription resources.

## Example Usage

```terraform
list "aws_pricingplanmanager_subscription" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
