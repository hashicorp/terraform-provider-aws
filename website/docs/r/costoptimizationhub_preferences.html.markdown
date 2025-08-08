---
subcategory: "Cost Optimization Hub"
layout: "aws"
page_title: "AWS: aws_costoptimizationhub_preferences"
description: |-
  Terraform resource for managing AWS Cost Optimization Hub Preferences.
---

# Resource: aws_costoptimizationhub_preferences

Terraform resource for managing AWS Cost Optimization Hub Preferences.

## Example Usage

### Basic Usage

```terraform
resource "aws_costoptimizationhub_preferences" "example" {
}
```

### Usage with all the arguments

```terraform
resource "aws_costoptimizationhub_preferences" "example" {
  member_account_discount_visibility = "None"
  savings_estimation_mode            = "AfterDiscounts"
}
```

## Argument Reference

The following arguments are optional:

* `member_account_discount_visibility` - (Optional) Customize whether the member accounts can see the "After Discounts" savings estimates. Valid values are `All` and `None`. Default value is `All`.
* `savings_estimation_mode` - (Optional) Customize how estimated monthly savings are calculated. Valid values are `BeforeDiscounts` and `AfterDiscounts`. Default value is `BeforeDiscounts`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Unique identifier for the preferences resource. Since preferences are for the entire account, this will be the 12-digit account id.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Cost Optimization Hub Preferences using the `id`. For example:

```terraform
import {
  to = aws_costoptimizationhub_preferences.example
  id = "111222333444"
}
```

Using `terraform import`, import Cost Optimization Hub Preferences using the `id`. For example:

```console
% terraform import aws_costoptimizationhub_preferences.example 111222333444
```
