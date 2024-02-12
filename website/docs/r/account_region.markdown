---
subcategory: "Account Management"
layout: "aws"
page_title: "AWS: aws_account_region"
description: |-
  Enable (Opt-In) or Disable (Opt-Out) a particular Region for an AWS account 
---

# Resource: aws_account_region

Enable (Opt-In) or Disable (Opt-Out) a particular Region for an AWS account.

## Example Usage

```terraform
resource "aws_account_region" "test" {
  region_name    = "ap-southeast-3"
  enabled        = "true"
}
```

## Argument Reference

This resource supports the following arguments:

* `account_id` - (Optional) The ID of the target account when managing member accounts. Will manage current user's account by default if omitted. To use this parameter, the caller must be an identity in the organization's management account or a delegated administrator account. The specified account ID must also be a member account in the same organization. The organization must have all features enabled , and the organization must have trusted access enabled for the Account Management service, and optionally a delegated admin account assigned.
* `region_name` - (Required) The region name to manage.
* `enabled` - (Optional) Whether the region is enabled.  Defaults to `true`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `opt_status` - The region opt status.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import the account region.
For example to import the region for the current account:

```terraform
import {
  to = aws_account_region.test
  id = ",ap-southeast-3"
}
```

To import the region for an organization member account:

```terraform
import {
  to = aws_account_region.test
  id = "1234567890,ap-southeast-3"
}
```

Using `terraform import`, For example to import the region for the current account:

```console
% terraform import aws_account_region.test ,ap-southeast-3
```

To import the region for an organization member account:

```console
% terraform import aws_account_region.test 1234567890,ap-southeast-3
```
