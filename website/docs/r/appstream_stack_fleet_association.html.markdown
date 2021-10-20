---
subcategory: "AppStream"
layout: "aws"
page_title: "AWS: aws_appstream_stack_fleet_association"
description: |-
  Manages an AppStream Stack Fleet association.
---

# Resource: aws_appstream_stack_fleet_association

Manages an AppStream Stack Fleet association.

## Example Usage

```terraform
resource "aws_appstream_fleet" "example" {
  name          = "FLEET NAME"
  image_name    = "Amazon-AppStream2-Sample-Image-02-04-2019"
  instance_type = "stream.standard.small"

  compute_capacity {
    desired_instances = 1
  }
}

resource "aws_appstream_stack" "example" {
  name = "STACK NAME"
}

resource "aws_appstream_stack_fleet_association" "example" {
  fleet_name = aws_appstream_fleet.example.name
  stack_name = aws_appstream_stack.example.name
}
```

## Argument Reference

The following arguments supported:

* `fleet_name` - (Required) Name of the fleet.
* `stack_name` (Optional) Name of the stack.

## Import

AppStream Stack Fleet Association can be imported by using the `stack_name` and `fleet_name` separated by a comma (`/`), e.g.,

```
$ terraform import aws_appstream_stack_fleet_association.example stackName/fleetName
```
