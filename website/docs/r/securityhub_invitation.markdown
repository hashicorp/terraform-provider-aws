---
layout: "aws"
page_title: "AWS: aws_securityhub_invitation"
sidebar_current: "docs-aws-resource-securityhub-invitation"
description: |-
  Provides a Security Hub invitation resource.
---

# aws_securityhub_invitation

Provides a Security Hub invitation resource.

## Example Usage

```hcl
resource "aws_securityhub_member" "example" {
  account_id = "123456789012"
  email      = "example@example.com"
}

resource "aws_securityhub_invitation" "example" {
  account_id = "123456789012"
}
```

## Argument Reference

The following arguments are supported:

* `account_id` - (Required) The ID of the invitee AWS account.

## Attributes Reference

The following attributes are exported in addition to the arguments listed above:

* `id` - The ID of the invitee AWS account (matches `account_id`).

## Import

Security Hub invitations can be imported using their account ID, e.g.

```
$ terraform import aws_securityhub_invitation.example 123456789012
```
