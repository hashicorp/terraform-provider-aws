---
layout: "aws"
page_title: "AWS: aws_guardduty_member"
sidebar_current: "docs-aws-resource-guardduty-member"
description: |-
  Provides a resource to manage a GuardDuty member
---

# aws_guardduty_member

Provides a resource to manage a GuardDuty member.

~> **NOTE:** Currently after using this resource, you must manually invite and accept member account invitations before GuardDuty will begin sending cross-account events. More information for how to accomplish this via the AWS Console or API can be found in the [GuardDuty User Guide](https://docs.aws.amazon.com/guardduty/latest/ug/guardduty_accounts.html). Terraform implementation of member invitation and acceptance resources can be tracked in [Github](https://github.com/terraform-providers/terraform-provider-aws/issues/2489).

## Example Usage

```hcl
resource "aws_guardduty_detector" "master" {
  enable = true
}

resource "aws_guardduty_detector" "member" {
  provider = "aws.dev"

  enable = true
}

resource "aws_guardduty_member" "member" {
  account_id  = "${aws_guardduty_detector.member.account_id}"
  detector_id = "${aws_guardduty_detector.master.id}"
  email       = "required@example.com"
}
```

## Argument Reference

The following arguments are supported:

* `account_id` - (Required) AWS account ID for member account.
* `detector_id` - (Required) The detector ID of the GuardDuty account where you want to create member accounts.
* `email` - (Required) Email address for member account.

## Attributes Reference

The following additional attributes are exported:

* `id` - The ID of the GuardDuty member

## Import

GuardDuty members can be imported using the the master GuardDuty detector ID and member AWS account ID, e.g.

```
$ terraform import aws_guardduty_member.MyMember 00b00fd5aecc0ab60a708659477e9617:123456789012
```
