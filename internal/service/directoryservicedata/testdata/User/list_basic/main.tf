# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_directoryservicedata_user" "test" {
  count = var.resource_count

  directory_id     = aws_directory_service_directory.test.id
  sam_account_name = "${var.samAccountNamePrefix}-${count.index}"
}

resource "aws_directory_service_directory" "test" {
  edition                      = "Standard"
  name                         = var.directoryDomain
  password                     = "SuperSecretPassw0rd"
  type                         = "MicrosoftAD"
  enable_directory_data_access = true

  vpc_settings {
    subnet_ids = aws_subnet.test[*].id
    vpc_id     = aws_vpc.test.id
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test" {
  count = 2

  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
}

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

variable "resource_count" {
  description = "Number of resources to create"
  type        = number
  nullable    = false
}

variable "directoryDomain" {
  type     = string
  nullable = false
}

variable "samAccountNamePrefix" {
  type     = string
  nullable = false
}
