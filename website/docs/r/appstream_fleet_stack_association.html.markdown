---
subcategory: "AppStream"
layout: "aws"
page_title: "AWS: aws_appstream_fleet_stack_association"
description: |-
  Manages an AppStream Fleet Stack association.
---

# Resource: aws_appstream_fleet_stack_association

Manages an AppStream Fleet Stack association.

## Example Usage

```terraform
resource "aws_appstream_fleet" "example" {
  name          = "NAME"
  image_name    = "Amazon-AppStream2-Sample-Image-02-04-2019"
  instance_type = "stream.standard.small"

  compute_capacity {
    desired_instances = 1
  }
}

resource "aws_appstream_stack" "example" {
  name = "STACK NAME"
}

resource "aws_appstream_fleet_stack_association" "example" {
  fleet_name = aws_appstream_fleet.example.name
  stack_name = aws_appstream_stack.example.name
}
```

## Argument Reference

The following arguments are required:

* `fleet_name` - (Required) Name of the fleet.
* `stack_name` (Required) Name of the stack.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Unique ID of the appstream stack fleet association, composed of the `fleet_name` and `stack_name` separated by a slash (`/`).

## Import

AppStream Stack Fleet Association can be imported by using the `fleet_name` and `stack_name` separated by a slash (`/`), e.g.,

```
$ terraform import aws_appstream_fleet_stack_association.example fleetName/stackName
```
