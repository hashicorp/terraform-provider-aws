---
subcategory: "Connect"
layout: "aws"
page_title: "AWS: aws_connect_queue"
description: |-
  Provides details about a specific Amazon Connect Queue.
---

# Data Source: aws_connect_queue

Provides details about a specific Amazon Connect Queue.

## Example Usage

By `name`

```hcl
data "aws_connect_queue" "example" {
  instance_id = "aaaaaaaa-bbbb-cccc-dddd-111111111111"
  name        = "Example"
}
```

By `queue_id`

```hcl
data "aws_connect_queue" "example" {
  instance_id = "aaaaaaaa-bbbb-cccc-dddd-111111111111"
  queue_id    = "cccccccc-bbbb-cccc-dddd-111111111111"
}
```

## Argument Reference

~> **NOTE:** `instance_id` and one of either `name` or `queue_id` is required.

The following arguments are supported:

* `queue_id` - (Optional) Returns information on a specific Queue by Queue id
* `instance_id` - (Required) Reference to the hosting Amazon Connect Instance
* `name` - (Optional) Returns information on a specific Queue by name

## Attributes Reference

In addition to all of the arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name (ARN) of the Queue.
* `description` - Specifies the description of the Queue.
* `hours_of_operation_id` - Specifies the identifier of the Hours of Operation.
* `id` - The identifier of the hosting Amazon Connect Instance and identifier of the Queue separated by a colon (`:`).
* `max_contacts` - Specifies the maximum number of contacts that can be in the queue before it is considered full. Minimum value of 0.
* `outbound_caller_config` - A block that defines the outbound caller ID name, number, and outbound whisper flow. The Outbound Caller Config block is documented below.
* `queue_id` - The identifier for the Queue.
* `status` - Specifies the description of the Queue. Values are `ENABLED` or `DISABLED`.
* `tags` - A map of tags assigned to the Queue.

A `outbound_caller_config` block supports the following arguments:

* `outbound_caller_id_name` - Specifies the caller ID name.
* `outbound_caller_id_number_id` - Specifies the caller ID number.
* `outbound_flow_id` - Specifies the outbound whisper flow to be used during an outbound call.
