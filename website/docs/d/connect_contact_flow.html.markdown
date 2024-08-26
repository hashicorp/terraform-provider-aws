---
subcategory: "Connect"
layout: "aws"
page_title: "AWS: aws_connect_contact_flow"
description: |-
  Provides details about a specific Amazon Connect Contact Flow.
---

# Data Source: aws_connect_contact_flow

Provides details about a specific Amazon Connect Contact Flow.

## Example Usage

By name

```hcl
data "aws_connect_contact_flow" "test" {
  instance_id = "aaaaaaaa-bbbb-cccc-dddd-111111111111"
  name        = "Test"
}
```

By contact_flow_id

```hcl
data "aws_connect_contact_flow" "test" {
  instance_id     = "aaaaaaaa-bbbb-cccc-dddd-111111111111"
  contact_flow_id = "cccccccc-bbbb-cccc-dddd-111111111111"
}
```

## Argument Reference

~> **NOTE:** `instance_id` and one of either `name` or `contact_flow_id` is required.

This data source supports the following arguments:

* `contact_flow_id` - (Optional) Returns information on a specific Contact Flow by contact flow id
* `instance_id` - (Required) Reference to the hosting Amazon Connect Instance
* `name` - (Optional) Returns information on a specific Contact Flow by name

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Contact Flow.
* `content` - Logic of the Contact Flow.
* `description` - Description of the Contact Flow.
* `tags` - Tags to assign to the Contact Flow.
* `type` - Type of Contact Flow.
