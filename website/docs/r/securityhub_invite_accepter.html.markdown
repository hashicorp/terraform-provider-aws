---
subcategory: "Security Hub"
layout: "aws"
page_title: "AWS: aws_securityhub_invite_accepter"
description: |-
  Accepts a Security Hub invitation.
---

# Resource: aws_securityhub_invite_accepter

-> **Note:** AWS accounts can only be associated with a single Security Hub master account. Destroying this resource will disassociate the member account from the master account.

Accepts a Security Hub invitation.

## Example Usage

```terraform
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
  depends_on = [aws_securityhub_account.invitee]
  master_id  = aws_securityhub_member.example.master_id
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `master_id` - (Required) The account ID of the master Security Hub account whose invitation you're accepting.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `invitation_id` - The ID of the invitation.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Security Hub invite acceptance using the account ID. For example:

```terraform
import {
  to = aws_securityhub_invite_accepter.example
  id = "123456789012"
}
```

Using `terraform import`, import Security Hub invite acceptance using the account ID. For example:

```console
% terraform import aws_securityhub_invite_accepter.example 123456789012
```
