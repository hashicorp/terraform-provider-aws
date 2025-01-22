---
subcategory: "Connect"
layout: "aws"
page_title: "AWS: aws_connect_routing_profile"
description: |-
  Provides details about a specific Amazon Connect Routing Profile.
---

# Data Source: aws_connect_routing_profile

Provides details about a specific Amazon Connect Routing Profile.

## Example Usage

By `name`

```hcl
data "aws_connect_routing_profile" "example" {
  instance_id = "aaaaaaaa-bbbb-cccc-dddd-111111111111"
  name        = "Example"
}
```

By `routing_profile_id`

```hcl
data "aws_connect_routing_profile" "example" {
  instance_id        = "aaaaaaaa-bbbb-cccc-dddd-111111111111"
  routing_profile_id = "cccccccc-bbbb-cccc-dddd-111111111111"
}
```

## Argument Reference

~> **NOTE:** `instance_id` and one of either `name` or `routing_profile_id` is required.

This data source supports the following arguments:

* `instance_id` - Reference to the hosting Amazon Connect Instance
* `name` - (Optional) Returns information on a specific Routing Profile by name
* `routing_profile_id` - (Optional) Returns information on a specific Routing Profile by Routing Profile id

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Routing Profile.
* `default_outbound_queue_id` - Specifies the default outbound queue for the Routing Profile.
* `description` - Description of the Routing Profile.
* `id` - Identifier of the hosting Amazon Connect Instance and identifier of the Routing Profile separated by a colon (`:`).
* `media_concurrencies` - One or more `media_concurrencies` blocks that specify the channels that agents can handle in the Contact Control Panel (CCP) for this Routing Profile. The `media_concurrencies` block is documented below.
* `queue_configs` - One or more `queue_configs` blocks that specify the inbound queues associated with the routing profile. If no queue is added, the agent only can make outbound calls. The `queue_configs` block is documented below.
* `tags` - Map of tags to assign to the Routing Profile.

A `media_concurrencies` block supports the following attributes:

* `channel` - Channels that agents can handle in the Contact Control Panel (CCP). Valid values are `VOICE`, `CHAT`, `TASK`.
* `concurrency` - Number of contacts an agent can have on a channel simultaneously. Valid Range for `VOICE`: Minimum value of 1. Maximum value of 1. Valid Range for `CHAT`: Minimum value of 1. Maximum value of 10. Valid Range for `TASK`: Minimum value of 1. Maximum value of 10.

A `queue_configs` block supports the following attributes:

* `channel` - Channels agents can handle in the Contact Control Panel (CCP) for this routing profile. Valid values are `VOICE`, `CHAT`, `TASK`.
* `delay` - Delay, in seconds, that a contact should be in the queue before they are routed to an available agent
* `priority` - Order in which contacts are to be handled for the queue.
* `queue_arn` - ARN for the queue.
* `queue_id` - Identifier for the queue.
* `queue_name` - Name for the queue.
