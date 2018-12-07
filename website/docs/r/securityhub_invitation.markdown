---
layout: "aws"
page_title: "AWS: aws_securityhub_invitation"
sidebar_current: "docs-aws-resource-securityhub-invitation"
description: |-
  Provides a Security Hub invitation resource.
---

# aws_securityhub_invitation

-> **Note:** The Security Hub API does not provide a way to revoke an invitation explicitly. Destroying this resource will remove the member account instead - this may cause a dirty plan on the next apply.

Provides a Security Hub invitation resource.

## Example Usage

```hcl
resource "aws_securityhub_account" "example" {}

resource "aws_securityhub_member" "example" {
  account_id = "123456789012"
  email      = "example@example.com"
}

resource "aws_securityhub_invitation" "example" {
  account_id = "${aws_securityhub_member.example.account_id}"
}
```

## Argument Reference

The following arguments are supported:

* `account_id` - (Required) The ID of the invitee AWS account.

## Attributes Reference

The following attributes are exported in addition to the arguments listed above:

* `id` - The ID of the invitee AWS account (matches `account_id`).
* `master_id` - The ID of the inviter AWS account.

## Import

Security Hub invitations can be imported using their account ID, e.g.

```
$ terraform import aws_securityhub_invitation.example 123456789012
```
