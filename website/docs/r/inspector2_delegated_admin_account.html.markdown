---
subcategory: "Inspector V2"
layout: "aws"
page_title: "AWS: aws_inspector2_delegated_admin_account"
description: |-
  Terraform resource for managing an AWS Inspector V2 Delegated Admin Account.
---

# Resource: aws_inspector2_delegated_admin_account

Terraform resource for managing an AWS Inspector V2 Delegated Admin Account.

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

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `relationship_status` - Status of this delegated admin account.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `15m`)
* `delete` - (Default `15m`)

## Import

Inspector V2 Delegated Admin Account can be imported using the `account_id`, e.g.,

```
$ terraform import aws_inspector2_delegated_admin_account.example 012345678901
```
