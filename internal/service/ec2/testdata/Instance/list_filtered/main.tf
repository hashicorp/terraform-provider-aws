# Copyright IBM Corp. 2014, 2025
# SPDX-License-Identifier: MPL-2.0

provider "aws" {}

resource "aws_instance" "expected" {
  count = 2

  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-arm64.id
  instance_type = "t4g.nano"

  metadata_options {
    http_tokens = "required"
  }

  tags = {
    Name = "expected-${count.index}"
  }
}

resource "aws_ec2_instance_state" "expected" {
  count = 2

  instance_id = aws_instance.expected[count.index].id
  state       = "stopped"
}

resource "aws_instance" "not_expected" {
  count = 2

  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-arm64.id
  instance_type = "t4g.nano"

  metadata_options {
    http_tokens = "required"
  }

  tags = {
    Name = "not-expected-${count.index}"
  }
}

# acctest.ConfigLatestAmazonLinux2HVMEBSARM64AMI

# acctest.configLatestAmazonLinux2HVMEBSAMI("arm64")

data "aws_ami" "amzn2-ami-minimal-hvm-ebs-arm64" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn2-ami-minimal-hvm-*"]
  }

  filter {
    name   = "root-device-type"
    values = ["ebs"]
  }

  filter {
    name   = "architecture"
    values = ["arm64"]
  }
}
