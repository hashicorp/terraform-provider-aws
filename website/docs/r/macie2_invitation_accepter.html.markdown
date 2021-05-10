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
  provider   = "awsalternate"
  account_id = "ACCOUNT ID"
  email      = "EMAIL"
  depends_on = [aws_macie2_account.primary]
}

resource "aws_macie2_invitation" "primary" {
  provider    = "awsalternate"
  account_ids = ["ACCOUNT IDS"]
  depends_on  = [aws_macie2_member.primary]
}

resource "aws_macie2_invitation_accepter" "test" {
  administrator_account_id = "ADMINISTRATOR ACCOUNT ID"
  depends_on               = [aws_macie2_invitation.primary]
}
```

## Argument Reference

The following arguments are supported:

* `administrator_account_id` - (Required) The AWS account ID for the account that sent the invitation.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The unique identifier (ID) of the macie invitation accepter.
* `invitation_id` - The unique identifier for the invitation.

## Import

`aws_macie2_invitation_accepter` can be imported using the id, e.g.

```
$ terraform import aws_macie2_invitation_accepter.example abcd1
```
