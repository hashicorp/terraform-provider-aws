---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_network_interface_sg_attachment"
description: |-
  Associates a security group with a network interface.
---

# Resource: aws_network_interface_sg_attachment

This resource attaches a security group to an Elastic Network Interface (ENI).
It can be used to attach a security group to any existing ENI, be it a
secondary ENI or one attached as the primary interface on an instance.

~> **NOTE on instances, interfaces, and security groups:** Terraform currently
provides the capability to assign security groups via the [`aws_instance`][1]
and the [`aws_network_interface`][2] resources. Using this resource in
conjunction with security groups provided in-line in those resources will cause
conflicts, and will lead to spurious diffs and undefined behavior - please use
one or the other.

[1]: /docs/providers/aws/d/instance.html
[2]: /docs/providers/aws/r/network_interface.html

## Example Usage

The following provides a very basic example of setting up an instance (provided
by `instance`) in the default security group, creating a security group
(provided by `sg`) and then attaching the security group to the instance's
primary network interface via the `aws_network_interface_sg_attachment` resource,
named `sg_attachment`:

```terraform
data "aws_ami" "ami" {
  most_recent = true

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*"]
  }

  owners = ["amazon"]
}

resource "aws_instance" "instance" {
  instance_type = "t2.micro"
  ami           = data.aws_ami.ami.id

  tags = {
    type = "terraform-test-instance"
  }
}

resource "aws_security_group" "sg" {
  tags = {
    type = "terraform-test-security-group"
  }
}

resource "aws_network_interface_sg_attachment" "sg_attachment" {
  security_group_id    = aws_security_group.sg.id
  network_interface_id = aws_instance.instance.primary_network_interface_id
}
```

In this example, `instance` is provided by the `aws_instance` data source,
fetching an external instance, possibly not managed by Terraform.
`sg_attachment` then attaches to the output instance's `network_interface_id`:

```terraform
data "aws_instance" "instance" {
  instance_id = "i-1234567890abcdef0"
}

resource "aws_security_group" "sg" {
  tags = {
    type = "terraform-test-security-group"
  }
}

resource "aws_network_interface_sg_attachment" "sg_attachment" {
  security_group_id    = aws_security_group.sg.id
  network_interface_id = data.aws_instance.instance.network_interface_id
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `security_group_id` - (Required) The ID of the security group.
* `network_interface_id` - (Required) The ID of the network interface to attach to.

## Attribute Reference

This resource exports no additional attributes.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `3m`)
- `read` - (Default `3m`)
- `delete` - (Default `3m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Network Interface Security Group attachments using the associated network interface ID and security group ID, separated by an underscore (`_`). For example:

```terraform
import {
  to = aws_network_interface_sg_attachment.sg_attachment
  id = "eni-1234567890abcdef0_sg-1234567890abcdef0"
}
```

Using `terraform import`, import Network Interface Security Group attachments using the associated network interface ID and security group ID, separated by an underscore (`_`). For example:

```console
% terraform import aws_network_interface_sg_attachment.sg_attachment eni-1234567890abcdef0_sg-1234567890abcdef0
```
