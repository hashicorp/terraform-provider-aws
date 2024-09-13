---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_ec2_instance_state"
description: |-
  Provides an EC2 instance state resource. This allows managing an instance power state. 
---

# Resource: aws_ec2_instance_state

Provides an EC2 instance state resource. This allows managing an instance power state.

~> **NOTE on Instance State Management:** AWS does not currently have an EC2 API operation to determine an instance has finished processing user data. As a result, this resource can interfere with user data processing. For example, this resource may stop an instance while the user data script is in mid run.

## Example Usage

```terraform
data "aws_ami" "ubuntu" {
  most_recent = true

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-focal-20.04-amd64-server-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

  owners = ["099720109477"] # Canonical
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.ubuntu.id
  instance_type = "t3.micro"

  tags = {
    Name = "HelloWorld"
  }
}

resource "aws_ec2_instance_state" "test" {
  instance_id = aws_instance.test.id
  state       = "stopped"
}
```

## Argument Reference

The following arguments are required:

* `instance_id` - (Required) ID of the instance.
* `state` - (Required) - State of the instance. Valid values are `stopped`, `running`.

The following arguments are optional:

* `force` - (Optional) Whether to request a forced stop when `state` is `stopped`. Otherwise (_i.e._, `state` is `running`), ignored. When an instance is forced to stop, it does not flush file system caches or file system metadata, and you must subsequently perform file system check and repair. Not recommended for Windows instances. Defaults to `false`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - ID of the instance (matches `instance_id`).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `10m`)
* `update` - (Default `10m`)
* `delete` - (Default `1m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_ec2_instance_state` using the `instance_id` attribute. For example:

```terraform
import {
  to = aws_ec2_instance_state.test
  id = "i-02cae6557dfcf2f96"
}
```

Using `terraform import`, import `aws_ec2_instance_state` using the `instance_id` attribute. For example:

```console
% terraform import aws_ec2_instance_state.test i-02cae6557dfcf2f96
```
