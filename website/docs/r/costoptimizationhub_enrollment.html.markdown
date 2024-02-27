---
subcategory: "Cost Optimization Hub"
layout: "aws"
page_title: "AWS: aws_costoptimizationhub_enrollment"
description: |-
  Terraform resource for managing AWS Cost Optimization Hub Enrollment.
---

# Resource: aws_costoptimizationhub_enrollment

Terraform resource for managing AWS Cost Optimization Hub Enrollment.

## Example Usage

### Basic Usage

```terraform
resource "aws_costoptimizationhub_enrollment" "example" {
}
```

### Usage with all the arguments

```terraform
resource "aws_costoptimizationhub_enrollment" "example" {
  include_member_accounts = true
  member_account_discount_visibility = "None"
  savings_estimation_mode = "AfterDiscounts"
}
```

## Argument Reference

The following arguments are optional:

* `include_member_accounts` - (Optional) Flag to enroll member accounts of the organization if the account is the management account. Default value is `false`
* `member_account_discount_visibility` - (Optional) Possible values are `All` and `None`. Default is `All`.
* `savings_estimation_modekey` - (Optional) Possible values are `BeforeDiscounts` and `AfterDiscounts`. Default value is `BeforeDiscounts`

## Attribute Reference

This resource exports the following attribute in addition to the arguments above:

* `id` - Unique identifier for the enrollment. Since enrollment is for the entire account, this will be the 12-digit account id.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Cost Optimization Hub Enrollment using the `id`. For example:

```terraform
import {
  to = aws_costoptimizationhub_enrollment.example
  id = "111222333444"
}
```

Using `terraform import`, import Cost Optimization Hub Enrollment using the `id`. For example:

```console
% terraform import aws_costoptimizationhub_enrollment.example 111222333444
```
