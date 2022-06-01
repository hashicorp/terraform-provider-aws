---
subcategory: "Connect"
layout: "aws"
page_title: "AWS: aws_connect_queue"
description: |-
  Provides details about a specific Amazon Connect Queue
---

# Resource: aws_connect_queue

Provides an Amazon Connect Queue resource. For more information see
[Amazon Connect: Getting Started](https://docs.aws.amazon.com/connect/latest/adminguide/amazon-connect-get-started.html)

~> **NOTE:** Due to The behaviour of Amazon Connect you cannot delete queues.

## Example Usage

### Basic

```terraform
resource "aws_connect_queue" "test" {
  instance_id           = "aaaaaaaa-bbbb-cccc-dddd-111111111111"
  name                  = "Example Name"
  description           = "Example Description"
  hours_of_operation_id = "12345678-1234-1234-1234-123456789012"

  tags = {
    "Name" = "Example Queue",
  }
}
```

### With Quick Connect IDs

```terraform
resource "aws_connect_queue" "test" {
  instance_id           = "aaaaaaaa-bbbb-cccc-dddd-111111111111"
  name                  = "Example Name"
  description           = "Example Description"
  hours_of_operation_id = "12345678-1234-1234-1234-123456789012"

  quick_connect_ids = [
    "12345678-abcd-1234-abcd-123456789012"
  ]

  tags = {
    "Name" = "Example Queue with Quick Connect IDs",
  }
}
```

### With Outbound Caller Config

```terraform
resource "aws_connect_queue" "test" {
  instance_id           = "aaaaaaaa-bbbb-cccc-dddd-111111111111"
  name                  = "Example Name"
  description           = "Example Description"
  hours_of_operation_id = "12345678-1234-1234-1234-123456789012"

  outbound_caller_config {
    outbound_caller_id_name      = "example"
    outbound_caller_id_number_id = "12345678-abcd-1234-abcd-123456789012"
    outbound_flow_id             = "87654321-defg-1234-defg-987654321234"
  }

  tags = {
    "Name" = "Example Queue with Outbound Caller Config",
  }
}
```

## Argument Reference

The following arguments are supported:

* `description` - (Optional) Specifies the description of the Queue.
* `hours_of_operation_id` - (Required) Specifies the identifier of the Hours of Operation.
* `instance_id` - (Required) Specifies the identifier of the hosting Amazon Connect Instance.
* `max_contacts` - (Optional) Specifies the maximum number of contacts that can be in the queue before it is considered full. Minimum value of 0.
* `name` - (Required) Specifies the name of the Queue.
* `outbound_caller_config` - (Required) A block that defines the outbound caller ID name, number, and outbound whisper flow. The Outbound Caller Config block is documented below.
* `quick_connect_ids` - (Optional) Specifies a list of quick connects ids that determine the quick connects available to agents who are working the queue.
* `status` - (Optional) Specifies the description of the Queue. Valid values are `ENABLED`, `DISABLED`.
* `tags` - (Optional) Tags to apply to the Queue. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

A `outbound_caller_config` block supports the following arguments:

* `outbound_caller_id_name` - (Optional) Specifies the caller ID name.
* `outbound_caller_id_number_id` - (Optional) Specifies the caller ID number.
* `outbound_flow_id` - (Optional) Specifies outbound whisper flow to be used during an outbound call.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name (ARN) of the Queue.
* `queue_id` - The identifier for the Queue.
* `id` - The identifier of the hosting Amazon Connect Instance and identifier of the Queue separated by a colon (`:`).
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

Amazon Connect Queues can be imported using the `instance_id` and `queue_id` separated by a colon (`:`), e.g.,

```
$ terraform import aws_connect_queue.example f1288a1f-6193-445a-b47e-af739b2:c1d4e5f6-1b3c-1b3c-1b3c-c1d4e5f6c1d4e5
```
