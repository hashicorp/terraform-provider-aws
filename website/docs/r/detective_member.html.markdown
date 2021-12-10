---
subcategory: "Detective"
layout: "aws"
page_title: "AWS: aws_detective_member"
description: |-
  Provides a resource to manage an Amazon  Detective.
---

# Resource: aws_detective_member

Provides a resource to manage an [Amazon Detective](https://docs.aws.amazon.com/detective/latest/APIReference/API_CreateMembers.html).

## Example Usage

```terraform
resource "aws_detective_graph" "example" {}

resource "aws_detective_member" "example" {
  account_id                            = "AWS ACCOUNT ID"
  email                                 = "EMAIL"
  graph_arn                             = aws_detective_graph.example.id
  invitation_message                    = "Message of the invitation"
  invitation_disable_email_notification = true
}
```

## Argument Reference

The following arguments are supported:

* `account_id` - (Required) AWS account ID for the account.
* `email_address` - (Required) Email address for the account.
* `graph_arn` - (Required) ARN of the behavior graph to invite the member accounts to contribute their data to.
* `message` - (Optional) A custom message to include in the invitation. Amazon Detective adds this message to the standard content that it sends for an invitation.
* `disable_email_notification` - (Optional) Specifies whether to send an email notification to the root user of each account that the invitation will be sent to. This notification is in addition to an alert that the root user receives in AWS Personal Health Dashboard. To send an email notification to the root user of each account, set this value to `true`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Unique identifier (ID) of the Detective.
* `status` - Current membership status of the member account.
* `administrator_id` - AWS account ID for the administrator account.
* `volume_usage_in_bytes` - Data volume in bytes per day for the member account.
* `volume_usage_updated_time` - Date and time, in UTC and extended RFC 3339 format, of the member account data volume was last updated.
* `invited_time` - Date and time, in UTC and extended RFC 3339 format, when an Amazon Detective membership invitation was last sent to the account.
* `updated_time` - Date and time, in UTC and extended RFC 3339 format, of the most recent change.

## Import

`aws_detective_member` can be imported using the account ID of the member account and the arn of a graph, e.g.

```
$ terraform import aws_detective_member.example 123456789012/arn:aws:detective:us-east1:graph:testing
```