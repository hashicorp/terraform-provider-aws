---
subcategory: "Macie"
layout: "aws"
page_title: "AWS: aws_macie2_invitation"
description: |-
  Provides a resource to manage an Amazon Macie Invitation.
---

# Resource: aws_macie2_invitation

Provides a resource to manage an [Amazon Macie Invitation](https://docs.aws.amazon.com/macie/latest/APIReference/invitations.html).

## Example Usage

```terraform
resource "aws_macie2_account" "test" {}

resource "aws_macie2_member" "test" {
  account_id = "AWS ACCOUNT ID"
  email      = "EMAIL"
  depends_on = [aws_macie2_account.test]
}

resource "aws_macie2_invitation" "test" {
  account_ids = ["ACCOUNT IDS"]
  depends_on  = [aws_macie2_member.test]
}
```

## Argument Reference

The following arguments are supported:

* `account_ids` - (Required) An array that lists AWS account IDs, one for each account to send the invitation to.
* `disable_email_notification` - (Optional) Specifies whether to send an email notification to the root user of each account that the invitation will be sent to. This notification is in addition to an alert that the root user receives in AWS Personal Health Dashboard. To send an email notification to the root user of each account, set this value to `true`.
* `message` - (Optional) A custom message to include in the invitation. Amazon Macie adds this message to the standard content that it sends for an invitation.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The unique identifier (ID) of the macie invitation.
* `invited_at` - The date and time, in UTC and extended RFC 3339 format, when an Amazon Macie membership invitation was last sent to the account. This value is null if a Macie invitation hasn't been sent to the account.

## Import

`aws_macie2_invitation` can be imported using the id, e.g.

```
$ terraform import aws_macie2_invitation.example abcd1
```
