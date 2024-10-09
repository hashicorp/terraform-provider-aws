---
subcategory: "Inspector"
layout: "aws"
page_title: "AWS: aws_inspector2_delegated_admin_account"
description: |-
  Terraform resource for managing an Amazon Inspector Delegated Admin Account.
---

# Resource: aws_inspector2_delegated_admin_account

Terraform resource for managing an Amazon Inspector Delegated Admin Account.

## Example Usage

### Basic Usage

```terraform
data "aws_caller_identity" "current" {}

resource "aws_inspector2_delegated_admin_account" "example" {
  account_id = data.aws_caller_identity.current.account_id
}
```

## Argument Reference

The following arguments are required:

* `account_id` - (Required) Account to enable as delegated admin account.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `relationship_status` - Status of this delegated admin account.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `15m`)
* `delete` - (Default `15m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Inspector Delegated Admin Account using the `account_id`. For example:

```terraform
import {
  to = aws_inspector2_delegated_admin_account.example
  id = "012345678901"
}
```

Using `terraform import`, import Inspector Delegated Admin Account using the `account_id`. For example:

```console
% terraform import aws_inspector2_delegated_admin_account.example 012345678901
```
