---
subcategory: "Connect"
layout: "aws"
page_title: "AWS: aws_connect_hours_of_operation"
description: |-
  Provides details about a specific Amazon Connect Hours of Operation.
---

# Data Source: aws_connect_hours_of_operation

Provides details about a specific Amazon Connect Hours of Operation.

## Example Usage

By `name`

```hcl
data "aws_connect_hours_of_operation" "test" {
  instance_id = "aaaaaaaa-bbbb-cccc-dddd-111111111111"
  name        = "Test"
}
```

By `hours_of_operation_id`

```hcl
data "aws_connect_hours_of_operation" "test" {
  instance_id           = "aaaaaaaa-bbbb-cccc-dddd-111111111111"
  hours_of_operation_id = "cccccccc-bbbb-cccc-dddd-111111111111"
}
```

## Argument Reference

~> **NOTE:** `instance_id` and one of either `name` or `hours_of_operation_id` is required.

This data source supports the following arguments:

* `hours_of_operation_id` - (Optional) Returns information on a specific Hours of Operation by hours of operation id
* `instance_id` - (Required) Reference to the hosting Amazon Connect Instance
* `name` - (Optional) Returns information on a specific Hours of Operation by name

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Hours of Operation.
* `config` - Configuration information for the hours of operation: day, start time, and end time . Config blocks are documented below. Config blocks are documented below.
* `description` - Description of the Hours of Operation.
* `hours_of_operation_id` - The identifier for the hours of operation.
* `instance_id` - Identifier of the hosting Amazon Connect Instance.
* `name` - Name of the Hours of Operation.
* `tags` - Map of tags to assign to the Hours of Operation.
* `time_zone` - Time zone of the Hours of Operation.

A `config` block supports the following arguments:

* `day` - Day that the hours of operation applies to.
* `end_time` - End time block specifies the time that your contact center closes. The `end_time` is documented below.
* `start_time` - Start time block specifies the time that your contact center opens. The `start_time` is documented below.

A `end_time` block supports the following arguments:

* `hours` - Hour of closing.
* `minutes` - Minute of closing.

A `start_time` block supports the following arguments:

* `hours` - Hour of opening.
* `minutes` - Minute of opening.
