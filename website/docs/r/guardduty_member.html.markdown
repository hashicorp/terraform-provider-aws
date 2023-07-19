---
subcategory: "GuardDuty"
layout: "aws"
page_title: "AWS: aws_guardduty_member"
description: |-
  Provides a resource to manage a GuardDuty member
---

# Resource: aws_guardduty_member

Provides a resource to manage a GuardDuty member. To accept invitations in member accounts, see the [`aws_guardduty_invite_accepter` resource](/docs/providers/aws/r/guardduty_invite_accepter.html).

## Example Usage

```terraform
resource "aws_guardduty_detector" "primary" {
  enable = true
}

resource "aws_guardduty_detector" "member" {
  provider = aws.dev

  enable = true
}

resource "aws_guardduty_member" "member" {
  account_id         = aws_guardduty_detector.member.account_id
  detector_id        = aws_guardduty_detector.primary.id
  email              = "required@example.com"
  invite             = true
  invitation_message = "please accept guardduty invitation"
}
```

## Argument Reference

This resource supports the following arguments:

* `account_id` - (Required) AWS account ID for member account.
* `detector_id` - (Required) The detector ID of the GuardDuty account where you want to create member accounts.
* `email` - (Required) Email address for member account.
* `invite` - (Optional) Boolean whether to invite the account to GuardDuty as a member. Defaults to `false`. To detect if an invitation needs to be (re-)sent, the Terraform state value is `true` based on a `relationship_status` of `Disabled`, `Enabled`, `Invited`, or `EmailVerificationInProgress`.
* `invitation_message` - (Optional) Message for invitation.
* `disable_email_notification` - (Optional) Boolean whether an email notification is sent to the accounts. Defaults to `false`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ID of the GuardDuty member
* `relationship_status` - The status of the relationship between the member account and its primary account. More information can be found in [Amazon GuardDuty API Reference](https://docs.aws.amazon.com/guardduty/latest/ug/get-members.html).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `1m`)
- `update` - (Default `1m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import GuardDuty members using the primary GuardDuty detector ID and member AWS account ID. For example:

```terraform
import {
  to = aws_guardduty_member.MyMember
  id = "00b00fd5aecc0ab60a708659477e9617:123456789012"
}
```

Using `terraform import`, import GuardDuty members using the primary GuardDuty detector ID and member AWS account ID. For example:

```console
% terraform import aws_guardduty_member.MyMember 00b00fd5aecc0ab60a708659477e9617:123456789012
```
