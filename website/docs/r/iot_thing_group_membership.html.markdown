---
subcategory: "IoT"
layout: "aws"
page_title: "AWS: aws_iot_thing_group_membership"
description: |-
    Adds an IoT Thing to an IoT Thing Group.
---

# Resource: aws_iot_thing_group_membership

Adds an IoT Thing to an IoT Thing Group.

## Example Usage

```terraform
resource "aws_iot_thing_group_membership" "example" {
  thing_name       = "example-thing"
  thing_group_name = "example-group"

  override_dynamic_group = true
}
```

## Argument Reference

* `thing_name` - (Required) The name of the thing to add to a group.
* `thing_group_name` - (Required) The name of the group to which you are adding a thing.
* `override_dynamic_group` - (Optional) Override dynamic thing groups with static thing groups when 10-group limit is reached. If a thing belongs to 10 thing groups, and one or more of those groups are dynamic thing groups, adding a thing to a static group removes the thing from the last dynamic group.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The membership ID.

## Import

IoT Thing Group Membership can be imported using the thing group name and thing name.

```
$ terraform import aws_iot_thing_group_membership.example thing_group_name/thing_name
```
