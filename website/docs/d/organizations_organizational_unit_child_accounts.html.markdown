---
subcategory: "Organizations"
layout: "aws"
page_title: "AWS: aws_organizations_organizational_unit_child_accounts"
description: |-
  Get all direct child accounts under a parent organizational unit. This only provides immediate children, not all children.
---

# Data Source: aws_organizations_organizational_unit_child_accounts

Get all direct child accounts under a parent organizational unit. This only provides immediate children, not all children.

## Example Usage

```terraform
data "aws_organizations_organization" "org" {}

data "aws_organizations_organizational_unit_child_accounts" "accounts" {
  parent_id = data.aws_organizations_organization.org.roots[0].id
}
```

## Argument Reference

* `parent_id` - (Required) The parent ID of the accounts.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `accounts` - List of child accounts, which have the following attributes:
    * `arn` - The Amazon Resource Name (ARN) of the account.
    * `email` - The email address associated with the AWS account.
    * `id` - The unique identifier (ID) of the account.
    * `name` - The friendly name of the account.
    * `status` - The status of the account in the organization.
* `id` - Parent identifier of the organizational units.
