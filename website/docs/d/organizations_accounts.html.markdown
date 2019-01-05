---
layout: "aws"
page_title: "AWS: aws_organizations_accounts"
sidebar_current: "docs-aws-datasource-organizations-accounts"
description: |-
  Get account IDs associated with an AWS Organization account.
---

# Data Source: aws_organizations_accounts

Use this data source to get the account IDs associated with a particular AWS Organization account.

## Example Usage

```hcl
data "aws_organizations_accounts" "current" {}

output "account_id_0" {
  value = "${data.aws_caller_identity.current.account_id[0]}"
}

output "account_id_1" {
  value = "${data.aws_caller_identity.current.account_id[1]}"
}
```

## Argument Reference

There are no arguments available for this data source.

## Attributes Reference

* `account_ids` - A list of account ID numbers from the account that manages the AWS Organization configuration.
