# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-arm64.id
  instance_type = "t4g.nano"

  metadata_options {
    http_tokens = "required"
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

