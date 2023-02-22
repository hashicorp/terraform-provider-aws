---
subcategory: "Lightsail"
layout: "aws"
page_title: "AWS: aws_lightsail_disk_attachment"
description: |-
  Attaches a Lightsail disk to a Lightsail Instance
---

# Resource: aws_lightsail_disk_attachment

Attaches a Lightsail disk to a Lightsail Instance

## Example Usage

```terraform
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_lightsail_disk" "test" {
  name              = "test-disk"
  size_in_gb        = 8
  availability_zone = data.aws_availability_zones.available.names[0]
}

resource "aws_lightsail_instance" "test" {
  name              = "test-instance"
  availability_zone = data.aws_availability_zones.available.names[0]
  blueprint_id      = "amazon_linux"
  bundle_id         = "nano_1_0"
}

resource "aws_lightsail_disk_attachment" "test" {
  disk_name     = aws_lightsail_disk.test.name
  instance_name = aws_lightsail_instance.test.name
  disk_path     = "/dev/xvdf"
}
```

## Argument Reference

The following arguments are supported:

* `disk_name` - (Required) The name of the Lightsail Disk.
* `instance_name` - (Required) The name of the Lightsail Instance to attach to.
* `disk_path` - (Required) The disk path to expose to the instance.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - A combination of attributes to create a unique id: `disk_name`,`instance_name`

## Import

`aws_lightsail_disk` can be imported by using the id attribute, e.g.,

```shell
$ terraform import aws_lightsail_disk_attachment.test test-disk,test-instance
```
