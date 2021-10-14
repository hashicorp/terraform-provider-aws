---
subcategory: "Organizations"
layout: "aws"
page_title: "AWS: aws_organizations_accounts"
description: |-
Get all the given accounts info (including tags) owned by the current organization that the user's account belongs to
---

# Data Source: aws_organizations_accounts

Get all the given accounts info (including tags) owned by the current organization that the user's account belongs to.
This datasource will be very useful in multi-account setups.

~> **Note:** Account info retrieval must be done from the organization's master account.

!> **WARNING:** For very large multi-account setup this will perform a large number of API requests.

## Example Usage

### List all given account arns for the organization

```terraform
data "aws_organizations_accounts" "example" {}

output "account_ids" {
  value = data.aws_organizations_accounts.example.accounts[*].arn
}
```

## Argument Reference

The following argument is supported:

* `account_ids` - (Required) The accounts ID to fetch info from.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `accounts` - List of fetched accounts. All elements have these attributes:
  * `arn` - ARN of the account
  * `email` - Email of the account
  * `id` - Identifier of the account
  * `name` - Name of the account
  * `status` - Current status of the account
  * `joined_method` - Method used to create the account
  * `joined_timestamp` - Account creation timestamp
  * `tags` - Key-value map of account tags.
