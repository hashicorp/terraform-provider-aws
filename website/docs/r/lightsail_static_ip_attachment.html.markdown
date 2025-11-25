---
subcategory: "Lightsail"
layout: "aws"
page_title: "AWS: aws_lightsail_static_ip_attachment"
description: |-
  Manages a Lightsail Static IP Attachment.
---

# Resource: aws_lightsail_static_ip_attachment

Manages a static IP address attachment - relationship between a Lightsail static IP and Lightsail instance.

Use this resource to attach a static IP address to a Lightsail instance to provide a consistent public IP address that persists across instance restarts.

~> **Note:** Lightsail is currently only supported in a limited number of AWS Regions, please see ["Regions and Availability Zones in Amazon Lightsail"](https://lightsail.aws.amazon.com/ls/docs/overview/article/understanding-regions-and-availability-zones-in-amazon-lightsail) for more details.

## Example Usage

```terraform
resource "aws_lightsail_static_ip" "example" {
  name = "example"
}

resource "aws_lightsail_instance" "example" {
  name              = "example"
  availability_zone = "us-east-1a"
  blueprint_id      = "ubuntu_20_04"
  bundle_id         = "nano_2_0"
}

resource "aws_lightsail_static_ip_attachment" "example" {
  static_ip_name = aws_lightsail_static_ip.example.id
  instance_name  = aws_lightsail_instance.example.id
}
```

## Argument Reference

The following arguments are required:

* `instance_name` - (Required) Name of the Lightsail instance to attach the IP to.
* `static_ip_name` - (Required) Name of the allocated static IP.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `ip_address` - Allocated static IP address.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_lightsail_static_ip_attachment` using the static IP name. For example:

```terraform
import {
  to = aws_lightsail_static_ip_attachment.example
  id = "example-static-ip"
}
```

Using `terraform import`, import `aws_lightsail_static_ip_attachment` using the static IP name. For example:

```console
% terraform import aws_lightsail_static_ip_attachment.example example-static-ip
```
