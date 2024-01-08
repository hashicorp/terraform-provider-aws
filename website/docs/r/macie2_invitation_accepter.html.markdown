---
subcategory: "Macie"
layout: "aws"
page_title: "AWS: aws_macie2_invitation_accepter"
description: |-
  Provides a resource to manage an Amazon Macie Invitation Accepter.
---

# Resource: aws_macie2_invitation_accepter

Provides a resource to manage an [Amazon Macie Invitation Accepter](https://docs.aws.amazon.com/macie/latest/APIReference/invitations-accept.html).

## Example Usage

```terraform
resource "aws_macie2_account" "primary" {
  provider = "awsalternate"
}

resource "aws_macie2_account" "member" {}

resource "aws_macie2_member" "primary" {
  provider           = "awsalternate"
  account_id         = "ACCOUNT ID"
  email              = "EMAIL"
  invite             = true
  invitation_message = "Message of the invite"
  depends_on         = [aws_macie2_account.primary]
}

resource "aws_macie2_invitation_accepter" "member" {
  administrator_account_id = "ADMINISTRATOR ACCOUNT ID"
  depends_on               = [aws_macie2_member.primary]
}
```

## Argument Reference

This resource supports the following arguments:

* `administrator_account_id` - (Required) The AWS account ID for the account that sent the invitation.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The unique identifier (ID) of the macie invitation accepter.
* `invitation_id` - The unique identifier for the invitation.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_macie2_invitation_accepter` using the admin account ID. For example:

```terraform
import {
  to = aws_macie2_invitation_accepter.example
  id = "123456789012"
}
```

Using `terraform import`, import `aws_macie2_invitation_accepter` using the admin account ID. For example:

```console
% terraform import aws_macie2_invitation_accepter.example 123456789012
```
