---
subcategory: "Compute Optimizer"
layout: "aws"
page_title: "AWS: aws_computeoptimizer_enrollment_status"
description: |-
  Manages AWS Compute Optimizer enrollment status.
---

# Resource: aws_computeoptimizer_enrollment_status

Manages AWS Compute Optimizer enrollment status.

## Example Usage

```terraform
resource "aws_computeoptimizer_enrollment_status" "example" {
  status = "Active"
}
```

## Argument Reference

This resource supports the following arguments:

* `include_member_accounts` - (Optional) Whether to enroll member accounts of the organization if the account is the management account of an organization. Default is `false`.
* `status` - (Required) The enrollment status of the account. Valid values: `Active`, `Inactive`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `number_of_member_accounts_opted_in` - The count of organization member accounts that are opted in to the service, if your account is an organization management account.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `update` - (Default `5m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import enrollment status using the account ID. For example:

```terraform
import {
  to = aws_computeoptimizer_enrollment_status.example
  id = "123456789012"
}
```

Using `terraform import`, import enrollment status using the account ID. For example:

```console
% terraform import aws_computeoptimizer_enrollment_status.example 123456789012
```
