---
layout: "aws"
page_title: "AWS: aws_greengrass_group"
description: |-
    Creates and manages an AWS IoT Greengrass Group
---

# Resource: aws_greengrass_group

## Example Usage

```hcl
resource "aws_greengrass_group" "test" {
  name = "test_group"
}
```

## Argument Reference

* `name` - (Required) The name of the group.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `group_id` - The id of the group
* `arn` - The ARN of the group

## Import

IoT Greengrass Groups can be imported using the `name`, e.g.

```
$ terraform import aws_greengrass_group.group <group_id>
```
