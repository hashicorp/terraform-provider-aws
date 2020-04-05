---
subcategory: "Security Hub"
layout: "aws"
page_title: "AWS: aws_securityhub_invite_accepter"
description: |-
  Accepts a Security Hub invitation.
---

# aws_securityhub_invite_accepter

-> **Note:** AWS accounts can only be associated with a single Security Hub master account. Destroying this resource will disassociate the member account from the master account.

Accepts a Security Hub invitation.

## Example Usage

```hcl
resource "aws_securityhub_account" "example" {}

resource "aws_securityhub_member" "example" {
  account_id = "123456789012"
  email      = "example@example.com"
  invite     = true
}

resource "aws_securityhub_account" "invitee" {
  provider = "aws.invitee"
}

resource "aws_securityhub_invite_accepter" "invitee" {
  provider   = "aws.invitee"
  depends_on = [aws_securityhub_account.accepter]
  master_id  = aws_securityhub_member.example.master_id
}
```

## Argument Reference

The following arguments are supported:

* `master_id` - (Required) The account ID of the master Security Hub account whose invitation you're accepting.

## Attributes Reference

The following attributes are exported in addition to the arguments listed above:

* `id` - Returns `securityhub-invite-accepter`.

## Import

Security Hub invite acceptance can be imported using the word `securityhub-invite-accepter`, e.g.

```
$ terraform import aws_securityhub_invite_acceptor.example securityhub-invite-accepter
```