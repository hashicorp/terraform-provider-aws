---
layout: "aws"
page_title: "AWS: aws_securityhub_invitation_accepter"
sidebar_current: "docs-aws-resource-securityhub-invitation-accepter"
description: |-
  Accepts a Security Hub invitation.
---

# aws_securityhub_invitation_accepter

-> **Note:** AWS accounts can only be associated with a single Security Hub master account. Destroying this resource will disassociate the member account from the master account.

Accepts a Security Hub invitation.

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

resource "aws_securityhub_account" "invitee" {
  provider = "aws.invitee"
}

resource "aws_securityhub_invitation_accepter" "invitee" {
  provider  = "aws.invitee"
  master_id = "${aws_securityhub_invitation.example.master_id}"
}
```

## Argument Reference

The following arguments are supported:

* `master_id` - (Optiona) The account ID of the master Security Hub account whose invitation you're accepting.

## Attributes Reference

The following attributes are exported in addition to the arguments listed above:

* `id` - Returns `securityhub-invitation-accepter`.

## Import

Security Hub invite acceptance can be imported using the word `securityhub-invitation-accepter`, e.g.

```
$ terraform import aws_securityhub_invitation_acceptor.example securityhub-invitation-accepter
```
