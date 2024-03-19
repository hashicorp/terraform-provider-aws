---
subcategory: "Cost Optimization Hub"
layout: "aws"
page_title: "AWS: aws_costoptimizationhub_enrollment_status"
description: |-
  Terraform resource for managing AWS Cost Optimization Hub Enrollment Status.
---

# Resource: aws_costoptimizationhub_enrollment_status

Terraform resource for managing AWS Cost Optimization Hub Enrollment Status.

## Example Usage

### Basic Usage

```terraform
resource "aws_costoptimizationhub_enrollment_status" "example" {
}
```

### Usage with all the arguments

```terraform
resource "aws_costoptimizationhub_enrollment_status" "example" {
  include_member_accounts = true
}
```

## Argument Reference

The following arguments are optional:

* `include_member_accounts` - (Optional) Flag to enroll member accounts of the organization if the account is the management account. Default value is `false`

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Unique identifier for the enrollment. Since enrollment is for the entire account, this will be the 12-digit account id.
* `status` - Status of enrollment. When the resource is present in Terraform, it's status will always be `Active`

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Cost Optimization Hub Enrollment Status using the `id`. For example:

```terraform
import {
  to = aws_costoptimizationhub_enrollment_status.example
  id = "111222333444"
}
```

Using `terraform import`, import Cost Optimization Hub Enrollment Status using the `id`. For example:

```console
% terraform import aws_costoptimizationhub_enrollment_status.example 111222333444
```
