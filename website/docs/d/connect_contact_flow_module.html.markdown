---
subcategory: "Connect"
layout: "aws"
page_title: "AWS: aws_connect_contact_flow_module"
description: |-
  Provides details about a specific Amazon Connect Contact Flow Module.
---

# Data Source: aws_connect_contact_flow_module

Provides details about a specific Amazon Connect Contact Flow Module.

## Example Usage

By `name`

```hcl
data "aws_connect_contact_flow_module" "example" {
  instance_id = "aaaaaaaa-bbbb-cccc-dddd-111111111111"
  name        = "example"
}
```

By `contact_flow_module_id`

```hcl
data "aws_connect_contact_flow_module" "example" {
  instance_id            = "aaaaaaaa-bbbb-cccc-dddd-111111111111"
  contact_flow_module_id = "cccccccc-bbbb-cccc-dddd-111111111111"
}
```

## Argument Reference

~> **NOTE:** `instance_id` and one of either `name` or `contact_flow_module_id` is required.

The following arguments are supported:

* `contact_flow_module_id` - (Optional) Returns information on a specific Contact Flow Module by contact flow module id
* `instance_id` - (Required) Reference to the hosting Amazon Connect Instance
* `name` - (Optional) Returns information on a specific Contact Flow Module by name

## Attributes Reference

In addition to all of the arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name (ARN) of the Contact Flow Module.
* `content` - Specifies the logic of the Contact Flow Module.
* `description` - Specifies the description of the Contact Flow Module.
* `tags` - A map of tags to assign to the Contact Flow Module.
* `state` - Specifies the type of Contact Flow Module Module. Values are either `ACTIVE` or `ARCHIVED`.
* `status` - The status of the Contact Flow Module Module. Values are either `PUBLISHED` or `SAVED`.
