---
subcategory: "Security Hub"
layout: "aws"
page_title: "AWS: aws_securityhub_member"
description: |-
  Provides a Security Hub member resource.
---

# Resource: aws_securityhub_member

Provides a Security Hub member resource.

## Example Usage

```terraform
resource "aws_securityhub_account" "example" {}

resource "aws_securityhub_member" "example" {
  depends_on = [aws_securityhub_account.example]
  account_id = "123456789012"
  email      = "example@example.com"
  invite     = true
}
```

## Argument Reference

This resource supports the following arguments:

* `account_id` - (Required) The ID of the member AWS account.
* `email` - (Optional) The email of the member AWS account.
* `invite` - (Optional) Boolean whether to invite the account to Security Hub as a member. Defaults to `false`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ID of the member AWS account (matches `account_id`).
* `master_id` - The ID of the master Security Hub AWS account.
* `member_status` - The status of the member account relationship.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Security Hub members using their account ID. For example:

```terraform
import {
  to = aws_securityhub_member.example
  id = "123456789012"
}
```

Using `terraform import`, import Security Hub members using their account ID. For example:

```console
% terraform import aws_securityhub_member.example 123456789012
```
