---
subcategory: "GuardDuty"
layout: "aws"
page_title: "AWS: aws_guardduty_invite_accepter"
description: |-
  Provides a resource to accept a pending GuardDuty invite on creation, ensure the detector has the correct primary account on read, and disassociate with the primary account upon removal.
---

# Resource: aws_guardduty_invite_accepter

Provides a resource to accept a pending GuardDuty invite on creation, ensure the detector has the correct primary account on read, and disassociate with the primary account upon removal.

## Example Usage

```terraform
provider "aws" {
  alias = "primary"
}

provider "aws" {
  alias = "member"
}

resource "aws_guardduty_invite_accepter" "member" {
  depends_on = [aws_guardduty_member.member]
  provider   = aws.member

  detector_id       = aws_guardduty_detector.member.id
  master_account_id = aws_guardduty_detector.primary.account_id
}

resource "aws_guardduty_member" "member" {
  provider    = aws.primary
  account_id  = aws_guardduty_detector.member.account_id
  detector_id = aws_guardduty_detector.primary.id
  email       = "required@example.com"
  invite      = true
}

resource "aws_guardduty_detector" "primary" {
  provider = aws.primary
}

resource "aws_guardduty_detector" "member" {
  provider = aws.member
}
```

## Argument Reference

This resource supports the following arguments:

* `detector_id` - (Required) The detector ID of the member GuardDuty account.
* `master_account_id` - (Required) AWS account ID for primary account.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - GuardDuty member detector ID

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `1m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_guardduty_invite_accepter` using the member GuardDuty detector ID. For example:

```terraform
import {
  to = aws_guardduty_invite_accepter.member
  id = "00b00fd5aecc0ab60a708659477e9617"
}
```

Using `terraform import`, import `aws_guardduty_invite_accepter` using the member GuardDuty detector ID. For example:

```console
% terraform import aws_guardduty_invite_accepter.member 00b00fd5aecc0ab60a708659477e9617
```
