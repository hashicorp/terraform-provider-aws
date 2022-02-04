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

The following arguments are supported:

* `hours_of_operation_id` - (Optional) Returns information on a specific Hours of Operation by hours of operation id
* `instance_id` - (Required) Reference to the hosting Amazon Connect Instance
* `name` - (Optional) Returns information on a specific Hours of Operation by name

## Attributes Reference

In addition to all of the arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name (ARN) of the Hours of Operation.
* `config` - Specifies configuration information for the hours of operation: day, start time, and end time . Config blocks are documented below. Config blocks are documented below.
* `description` - Specifies the description of the Hours of Operation.
* `hours_of_operation_arn` - (**Deprecated**) The Amazon Resource Name (ARN) of the Hours of Operation.
* `hours_of_operation_id` - The identifier for the hours of operation.
* `instance_id` - Specifies the identifier of the hosting Amazon Connect Instance.
* `name` - Specifies the name of the Hours of Operation.
* `tags` - A map of tags to assign to the Hours of Operation.
* `time_zone` - Specifies the time zone of the Hours of Operation.

A `config` block supports the following arguments:

* `day` - Specifies the day that the hours of operation applies to.
* `end_time` - A end time block specifies the time that your contact center closes. The `end_time` is documented below.
* `start_time` - A start time block specifies the time that your contact center opens. The `start_time` is documented below.

A `end_time` block supports the following arguments:

* `hours` - Specifies the hour of closing.
* `minutes` - Specifies the minute of closing.

A `start_time` block supports the following arguments:

* `hours` - Specifies the hour of opening.
* `minutes` - Specifies the minute of opening.
