---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_eip_association"
description: |-
  Provides an AWS EIP Association
---

# Resource: aws_eip_association

Provides an AWS EIP Association as a top level resource, to associate and disassociate Elastic IPs from AWS Instances and Network Interfaces.

~> **NOTE:** Do not use this resource to associate an EIP to `aws_lb` or `aws_nat_gateway` resources. Instead use the `allocation_id` available in those resources to allow AWS to manage the association, otherwise you will see `AuthFailure` errors.

~> **NOTE:** `aws_eip_association` is useful in scenarios where EIPs are either pre-existing or distributed to customers or users and therefore cannot be changed.

## Example Usage

```terraform
resource "aws_eip_association" "eip_assoc" {
  instance_id   = aws_instance.web.id
  allocation_id = aws_eip.example.id
}

resource "aws_instance" "web" {
  ami               = "ami-21f78e11"
  availability_zone = "us-west-2a"
  instance_type     = "t2.micro"

  tags = {
    Name = "HelloWorld"
  }
}

resource "aws_eip" "example" {
  domain = "vpc"
}
```

## Argument Reference

This resource supports the following arguments:

* `allocation_id` - (Optional, Forces new resource) ID of the associated Elastic IP.
  This argument is required despite being optional at the resource level due to legacy support for EC2-Classic networking.
* `allow_reassociation` - (Optional, Forces new resource) Whether to allow an Elastic IP address to be re-associated.
  Defaults to `true`.
* `instance_id` - (Optional, Forces new resource) ID of the instance.
  The instance must have exactly one attached network interface.
  You can specify either the instance ID or the network interface ID, but not both.
* `network_interface_id` - (Optional, Forces new resource) ID of the network interface.
  If the instance has more than one network interface, you must specify a network interface ID.
  You can specify either the instance ID or the network interface ID, but not both.
* `private_ip_address` - (Optional, Forces new resource) Primary or secondary private IP address to associate with the Elastic IP address.
  If no private IP address is specified, the Elastic IP address is associated with the primary private IP address.
* `public_ip` - (Optional, Forces new resource, **Deprecated** since [EC2-Classic netwworking has retired](https://aws.amazon.com/blogs/aws/ec2-classic-is-retiring-heres-how-to-prepare/)) Address of the associated Elastic IP.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - ID that represents the association of the Elastic IP address with an instance.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import EIP Assocations using their association IDs. For example:

```terraform
import {
  to = aws_eip_association.test
  id = "eipassoc-ab12c345"
}
```

Using `terraform import`, import EIP Assocations using their association IDs. For example:

```console
% terraform import aws_eip_association.test eipassoc-ab12c345
```
