---
subcategory: "IAM (Identity & Access Management)"
layout: "aws"
page_title: "AWS: aws_iam_account_alias"
description: |-
  Manages the account alias for the AWS Account.
---

# Resource: aws_iam_account_alias

-> **Note:** There is only a single account alias per AWS account.

Manages the account alias for the AWS Account.

## Example Usage

```terraform
resource "aws_iam_account_alias" "alias" {
  account_alias = "my-account-alias"
}
```

## Argument Reference

This resource supports the following arguments:

* `account_alias` - (Required) The account alias

## Attribute Reference

This resource exports no additional attributes.

## Import

Import the current Account Alias using the `account_alias`. For example:

```
$ terraform import aws_iam_account_alias.alias my-account-alias
```
