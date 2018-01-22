---
layout: "aws"
page_title: "AWS: aws_organization_members"
sidebar_current: "docs-aws-datasource-orgnization-members"
description: |-
  Get AWS Organizations account ID's
---

# aws_organization_members

Use this data source to lookup current AWS accounts in your AWS Organization

## Example Usage

```hcl
data "aws_organizations_members" "accounts" {
  account_id = "${data.aws_caller_identity.current.account_id}"
}
```

## Argument Reference

Hand the parent account ID as an argument to retrieve all member accounts

* `account_id` - (Required) The acount id of the parent account.

## Attributes Reference

`accounts` is a list of account ID's for all accounts under the supplied parent account ID.