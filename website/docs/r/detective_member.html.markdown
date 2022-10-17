---
subcategory: "Detective"
layout: "aws"
page_title: "AWS: aws_detective_member"
description: |-
  Provides a resource to manage an Amazon Detective member.
---

# Resource: aws_detective_member

Provides a resource to manage an [Amazon Detective Member](https://docs.aws.amazon.com/detective/latest/APIReference/API_CreateMembers.html).

## Example Usage

```terraform
resource "aws_detective_graph" "example" {}

resource "aws_detective_member" "example" {
  account_id                 = "AWS ACCOUNT ID"
  email_address              = "EMAIL"
  graph_arn                  = aws_detective_graph.example.id
  message                    = "Message of the invitation"
  disable_email_notification = true
}
```

## Argument Reference

The following arguments are supported:

* `account_id` - (Required) AWS account ID for the account.
* `email_address` - (Required) Email address for the account.
* `graph_arn` - (Required) ARN of the behavior graph to invite the member accounts to contribute their data to.
* `message` - (Optional) A custom message to include in the invitation. Amazon Detective adds this message to the standard content that it sends for an invitation.
* `disable_email_notification` - (Optional) If set to true, then the root user of the invited account will _not_ receive an email notification. This notification is in addition to an alert that the root user receives in AWS Personal Health Dashboard. By default, this is set to `false`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Unique identifier (ID) of the Detective.
* `status` - Current membership status of the member account.
* `administrator_id` - AWS account ID for the administrator account.
* `volume_usage_in_bytes` - Data volume in bytes per day for the member account.
* `invited_time` - Date and time, in UTC and extended RFC 3339 format, when an Amazon Detective membership invitation was last sent to the account.
* `updated_time` - Date and time, in UTC and extended RFC 3339 format, of the most recent change to the member account's status.

## Import

`aws_detective_member` can be imported using the ARN of the graph followed by the account ID of the member account, e.g.

```
$ terraform import aws_detective_member.example arn:aws:detective:us-east-1:123456789101:graph:231684d34gh74g4bae1dbc7bd807d02d/123456789012
```
