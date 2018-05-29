---
layout: "aws"
page_title: "AWS: aws_organizations_account"
sidebar_current: "docs-aws-datasource-organizations_account"
description: |-
  Get information on an AWS Account in the joined AWS Organization.
---

# Data Source: aws_organizations_account

Use this data source to get details on an AWS Account in the joined organization.

~> **NOTE:** Account management must be done from the organization's master account.

## Example Usage

```hcl
data "aws_caller_identity" "current" {}

data "aws_organizations_account" "dev" {
  account_id = "${data.aws_caller_identity.current.account_id}"
}

resource "aws_ssm_parameter" "accounts_dev_email" {
  name  = "/terraform-example/accounts/dev/email"
  type  = "String"
  value = "${data.aws_organizations_account.dev.email}"
}
```

## Argument Reference

The following arguments are supported:

* `account_id` - (Required) The ID of the AWS Account

## Attributes Reference

`id` is set to the ID of the AWS Account. In addition, the following attributes
are exported:

* `arn` - The ARN for this account.
* `email` - The email address associated with the AWS account.
* `joined_method` - The method by which the account joined the organization.
* `joined_timestamp` - The date the account became a part of the organization.
* `name` - The friendly name of the account.
* `status` - The status of the account in the organization.
