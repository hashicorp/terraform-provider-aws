---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_instance"
description: |-
  Provides an EC2 instance state resource. This allows managing an instance power state. 
---

# Resource: aws_instance_state

Provides an EC2 instance state resource. This allows managing an instance power state.

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

resource "aws_instance_state" "test" {
  instance_id = aws_instance.test.id
  state       = "stopped"
}
```

## Argument Reference

The following arguments are supported:

* `instance_id` - (Required) The ID of the instance.
* `state` - (Required) - The state of the instance. Valid Options: `stopped`, `running`.
* `force` - (Optional) Forces the instances to stop. The instances do not have an opportunity to flush file system caches or file system metadata. If you use this option, you must perform file system check and repair procedures. This option is not recommended for Windows instances.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the instance. (matches `instance_id`).

## Import

`aws_instance_state` can be imported by using the `instance_id` attribute, e.g.,

```
$ terraform import aws_instance_state.test i-02cae6557dfcf2f96
```
