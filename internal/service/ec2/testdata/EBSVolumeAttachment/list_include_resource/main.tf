# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_volume_attachment" "test" {
  count = var.resource_count

  device_name = local.device_names[count.index]
  volume_id   = aws_ebs_volume.test[count.index].id
  instance_id = aws_instance.test.id
}

locals {
  device_names = ["/dev/sdh", "/dev/sdi", "/dev/sdj", "/dev/sdk", "/dev/sdl"]
}

# acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI

# acctest.configLatestAmazonLinux2HVMEBSAMI("x86_64")

data "aws_ami" "amzn2-ami-minimal-hvm-ebs-x86_64" {
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
    values = ["x86_64"]
  }
}

# acctest.ConfigAvailableAZsNoOptInDefaultExclude

data "aws_availability_zones" "available" {
  exclude_zone_ids = local.default_exclude_zone_ids
  state            = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

locals {
  default_exclude_zone_ids = ["usw2-az4", "usgw1-az2"]
}

data "aws_ec2_instance_type_offering" "available" {
  filter {
    name   = "instance-type"
    values = ["t3.micro", "t2.micro"]
  }

  filter {
    name   = "location"
    values = [data.aws_availability_zones.available.names[0]]
  }

  location_type            = "availability-zone"
  preferred_instance_types = ["t3.micro", "t2.micro"]
}

resource "aws_instance" "test" {
  ami               = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  availability_zone = data.aws_availability_zones.available.names[0]
  instance_type     = data.aws_ec2_instance_type_offering.available.instance_type
}

resource "aws_ebs_volume" "test" {
  count = var.resource_count

  availability_zone = data.aws_availability_zones.available.names[0]
  size              = 1
}

variable "resource_count" {
  description = "Number of resources to create"
  type        = number
  nullable    = false
}
