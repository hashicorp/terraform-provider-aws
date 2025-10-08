---
subcategory: "Lightsail"
layout: "aws"
page_title: "AWS: aws_lightsail_disk_attachment"
description: |-
  Manages the attachment of a Lightsail disk to an instance.
---

# Resource: aws_lightsail_disk_attachment

Manages a Lightsail disk attachment. Use this resource to attach additional storage disks to your Lightsail instances for expanded storage capacity.

## Example Usage

```terraform
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_lightsail_disk" "example" {
  name              = "example-disk"
  size_in_gb        = 8
  availability_zone = data.aws_availability_zones.available.names[0]
}

resource "aws_lightsail_instance" "example" {
  name              = "example-instance"
  availability_zone = data.aws_availability_zones.available.names[0]
  blueprint_id      = "amazon_linux_2"
  bundle_id         = "nano_3_0"
}

resource "aws_lightsail_disk_attachment" "example" {
  disk_name     = aws_lightsail_disk.example.name
  instance_name = aws_lightsail_instance.example.name
  disk_path     = "/dev/xvdf"
}
```

## Argument Reference

This resource supports the following arguments:

* `disk_name` - (Required) Name of the Lightsail disk.
* `disk_path` - (Required) Disk path to expose to the instance.
* `instance_name` - (Required) Name of the Lightsail instance to attach to.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Combination of attributes to create a unique id: `disk_name`,`instance_name`.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_lightsail_disk_attachment` using the id attribute. For example:

```terraform
import {
  to = aws_lightsail_disk_attachment.example
  id = "example-disk,example-instance"
}
```

Using `terraform import`, import `aws_lightsail_disk_attachment` using the id attribute. For example:

```console
% terraform import aws_lightsail_disk_attachment.example example-disk,example-instance
```
