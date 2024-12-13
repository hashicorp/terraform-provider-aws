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
  blueprint_id      = "amazon_linux_2"
  bundle_id         = "nano_3_0"
}

resource "aws_lightsail_disk_attachment" "test" {
  disk_name     = aws_lightsail_disk.test.name
  instance_name = aws_lightsail_instance.test.name
  disk_path     = "/dev/xvdf"
}
```

## Argument Reference

This resource supports the following arguments:

* `disk_name` - (Required) The name of the Lightsail Disk.
* `instance_name` - (Required) The name of the Lightsail Instance to attach to.
* `disk_path` - (Required) The disk path to expose to the instance.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - A combination of attributes to create a unique id: `disk_name`,`instance_name`

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_lightsail_disk` using the id attribute. For example:

```terraform
import {
  to = aws_lightsail_disk_attachment.test
  id = "test-disk,test-instance"
}
```

Using `terraform import`, import `aws_lightsail_disk` using the id attribute. For example:

```console
% terraform import aws_lightsail_disk_attachment.test test-disk,test-instance
```
