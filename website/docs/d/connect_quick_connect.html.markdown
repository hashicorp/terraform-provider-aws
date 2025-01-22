---
subcategory: "Connect"
layout: "aws"
page_title: "AWS: aws_connect_quick_connect"
description: |-
  Provides details about a specific Amazon Connect Quick Connect.
---

# Data Source: aws_connect_quick_connect

Provides details about a specific Amazon Connect Quick Connect.

## Example Usage

By `name`

```hcl
data "aws_connect_quick_connect" "example" {
  instance_id = "aaaaaaaa-bbbb-cccc-dddd-111111111111"
  name        = "Example"
}
```

By `quick_connect_id`

```hcl
data "aws_connect_quick_connect" "example" {
  instance_id      = "aaaaaaaa-bbbb-cccc-dddd-111111111111"
  quick_connect_id = "cccccccc-bbbb-cccc-dddd-111111111111"
}
```

## Argument Reference

~> **NOTE:** `instance_id` and one of either `name` or `quick_connect_id` is required.

This data source supports the following arguments:

* `quick_connect_id` - (Optional) Returns information on a specific Quick Connect by Quick Connect id
* `instance_id` - (Required) Reference to the hosting Amazon Connect Instance
* `name` - (Optional) Returns information on a specific Quick Connect by name

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Quick Connect.
* `description` - Description of the Quick Connect.
* `id` - Identifier of the hosting Amazon Connect Instance and identifier of the Quick Connect separated by a colon (`:`).
* `quick_connect_config` - A block that defines the configuration information for the Quick Connect: `quick_connect_type` and one of `phone_config`, `queue_config`, `user_config` . The Quick Connect Config block is documented below.
* `quick_connect_id` - Identifier for the Quick Connect.
* `tags` - Map of tags to assign to the Quick Connect.

A `quick_connect_config` block contains the following arguments:

* `quick_connect_type` - Configuration type of the Quick Connect. Valid values are `PHONE_NUMBER`, `QUEUE`, `USER`.
* `phone_config` - Phone configuration of the Quick Connect. This is returned only if `quick_connect_type` is `PHONE_NUMBER`. The `phone_config` block is documented below.
* `queue_config` - Queue configuration of the Quick Connect. This is returned only if `quick_connect_type` is `QUEUE`. The `queue_config` block is documented below.
* `user_config` - User configuration of the Quick Connect. This is returned only if `quick_connect_type` is `USER`. The `user_config` block is documented below.

A `phone_config` block contains the following arguments:

* `phone_number` - Phone number in in E.164 format.

A `queue_config` block contains the following arguments:

* `contact_flow_id` - Identifier of the contact flow.
* `queue_id` - Identifier for the queue.

A `user_config` block contains the following arguments:

* `contact_flow_id` - Identifier of the contact flow.
* `user_id` - Identifier for the user.
