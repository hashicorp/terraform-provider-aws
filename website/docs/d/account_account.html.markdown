---
subcategory: "Account Management"
layout: "aws"
page_title: "AWS: aws_account_account"
description: |-
  Provides details about an AWS Account.
---

# Data Source: aws_account_account

Provides the account information about an AWS Account.

## Example Usage

### Basic Usage

```terraform
data "aws_account_account" "example" {
}
```

### Organization Management Account Usage

```terraform
data "aws_account_account" "example" {
  account_id = "123456789000"
}
```

## Argument Reference

The following arguments are optional:

* `account_id` - (Optional) The ID of the target account when managing member accounts. The caller must be an identity in the organization's management account or a delegated administrator account. It will return current user's account by default if omitted.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `account_name` - The name of the account.
* `account_created_date` - The date and time the account was created.
