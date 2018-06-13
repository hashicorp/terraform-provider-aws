---
layout: "aws"
page_title: "AWS: aws_cloudwatch_event_permission"
sidebar_current: "docs-aws-resource-cloudwatch-event-permission"
description: |-
  Provides a resource to create a CloudWatch Events permission to support cross-account events in the current account default event bus.
---

# aws_cloudwatch_event_permission

Provides a resource to create a CloudWatch Events permission to support cross-account events in the current account default event bus.

## Example Usage

```hcl
resource "aws_cloudwatch_event_permission" "DevAccountAccess" {
  principal    = "123456789012"
  statement_id = "DevAccountAccess"
}
```

## Argument Reference

The following arguments are supported:

* `principal` - (Required) The 12-digit AWS account ID that you are permitting to put events to your default event bus. Specify `*` to permit any account to put events to your default event bus.
* `statement_id` - (Required) An identifier string for the external account that you are granting permissions to.
* `action` - (Optional) The action that you are enabling the other account to perform. Defaults to `events:PutEvents`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The statement ID of the CloudWatch Events permission.

## Import

CloudWatch Events permissions can be imported using the statement ID, e.g.

```shell
$ terraform import aws_cloudwatch_event_permission.DevAccountAccess DevAccountAccess
```
