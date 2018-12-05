---
layout: "aws"
page_title: "AWS: aws_securityhub_member"
sidebar_current: "docs-aws-resource-securityhub-member"
description: |-
  Provides a Security Hub member resource.
---

# aws_securityhub_member

Provides a Security Hub member resource.

## Example Usage

```hcl
resource "aws_securityhub_member" "example" {
  account_id = "123456789012"
  email      = "example@example.com"
}
```

## Argument Reference

The following arguments are supported:

* `account_id` - (Required) The ID of the member AWS account.
* `email` - (Required) The email of the member AWS account.

## Attributes Reference

The following attributes are exported in addition to the arguments listed above:

* `id` - The ID of the member AWS account (matches `account_id`).

## Import

Security Hub members can be imported using their account ID, e.g.

```
$ terraform import aws_securityhub_member.example 123456789012
```
