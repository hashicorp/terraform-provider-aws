---
subcategory: "Organizations"
layout: "aws"
page_title: "AWS: aws_organizations_accounts"
description: |-
  Get all accounts belonging to an organizational unit.
---

# Data Source: aws_organizations_accounts

Get all accounts belonging to an organizational unit.

## Example Usage

```terraform
data "aws_organizations_organization" "org" {}

data "aws_organizations_accounts" "ou" {
  parent_id = data.aws_organizations_organization.org.roots[0].id
}
```

## Argument Reference

* `parent_id` - (Required) The parent ID of the accounts.

## Attributes Reference

* `children` - List of child accounts, which have the following attributes:
    * `arn` - ARN of the organizational account
    * `name` - Name of the organizational account
    * `id` - ID of the organizational account
    * `email` - Email of the account
    * `status` - Status of the account (either `ACTIVE` or `SUSPENDED`)
* `id` - Parent identifier of the accounts.
