---
subcategory: "EventBridge (CloudWatch Events)"
layout: "aws"
page_title: "AWS: aws_cloudwatch_event_bus_policy"
description: |-
  Provides a resource to create an EventBridge policy to support cross-account events.
---

# Resource: aws_cloudwatch_event_permission

Provides a resource to create an EventBridge resource policy to support cross-account events.

~> **Note:** EventBridge was formerly known as CloudWatch Events. The functionality is identical.

~> **Note:** The cloudwatch eventbus policy resource is incompatible with the cloudwatch event permissions resource and will overwrite them.

## Example Usage

### Account Access

```hcl
resource "aws_cloudwatch_event_bus_policy" "test" {
  policy         = data.aws_iam_policy_document.access.json
  event_bus_name = aws_cloudwatch_event_bus.test.name
}
```

## Argument Reference

The following arguments are supported:

* `policy` - (Required) The text of the policy. For more information about building AWS IAM policy documents with Terraform, see the [AWS IAM Policy Document Guide](https://learn.hashicorp.com/terraform/aws/iam-policy).
* `event_bus_name` - (Optional) The event bus to set the permissions on. If you omit this, the permissions are set on the `default` event bus.

## Import

EventBridge permissions can be imported using the `event_bus_name`, e.g.

```shell
$ terraform import aws_cloudwatch_event_bus_policy.DevAccountAccess example-event-bus
```
