# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_timestreaminfluxdb_db_backup" "test" {
  count  = var.resource_count
  region = var.region

  db_resource_id = aws_timestreaminfluxdb_db_instance.test.id
  name           = "${var.rName}-${count.index}"
}

resource "aws_timestreaminfluxdb_db_instance" "test" {
  region = var.region

  name                   = var.rName
  allocated_storage      = 20
  username               = "admin"
  password               = "testpassword"
  vpc_subnet_ids         = aws_subnet.test[*].id
  vpc_security_group_ids = [aws_security_group.test.id]
  db_instance_type       = "db.influx.medium"
  bucket                 = "initial"
  organization           = "organization"
}

resource "aws_vpc" "test" {
  region = var.region

  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test" {
  region = var.region

  count = 1

  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
}

data "aws_availability_zones" "available" {
  region = var.region

  exclude_zone_ids = ["usw2-az4", "usgw1-az2"]
  state            = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_security_group" "test" {
  region = var.region

  vpc_id = aws_vpc.test.id
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "resource_count" {
  description = "Number of resources to create"
  type        = number
  nullable    = false
}

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
