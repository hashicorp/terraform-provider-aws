---
subcategory: "AppStream 2.0"
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

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Unique ID of the appstream stack fleet association, composed of the `fleet_name` and `stack_name` separated by a slash (`/`).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import AppStream Stack Fleet Association using the `fleet_name` and `stack_name` separated by a slash (`/`). For example:

```terraform
import {
  to = aws_appstream_fleet_stack_association.example
  id = "fleetName/stackName"
}
```

Using `terraform import`, import AppStream Stack Fleet Association using the `fleet_name` and `stack_name` separated by a slash (`/`). For example:

```console
% terraform import aws_appstream_fleet_stack_association.example fleetName/stackName
```
