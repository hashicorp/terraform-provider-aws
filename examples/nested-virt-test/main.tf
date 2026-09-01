# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

terraform {
  required_version = ">= 0.12"
}

provider "aws" {
  region = var.aws_region
}

data "aws_ami" "amazon_linux" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["al2023-ami-*-x86_64"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}

# EC2 instance with nested virtualization enabled
resource "aws_instance" "nested_virt" {
  ami           = data.aws_ami.amazon_linux.id
  instance_type = var.instance_type

  cpu_options {
    nested_virtualization = "enabled"
  }

  tags = {
    Name = "nested-virt-example"
  }
}
