# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test" {
  count = 2

  vpc_id            = aws_vpc.test.id
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  availability_zone = data.aws_availability_zones.available.names[count.index]
}

resource "aws_security_group" "test" {
  count = 2

  vpc_id = aws_vpc.test.id
  name   = "test-sg-${count.index}"
}

data "aws_availability_zones" "available" {
  state = "available"
  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_workspacesweb_network_settings" "test" {
  vpc_id             = aws_vpc.test.id
  subnet_ids         = [aws_subnet.test[0].id, aws_subnet.test[1].id]
  security_group_ids = [aws_security_group.test[0].id, aws_security_group.test[1].id]

  tags = var.resource_tags

}

variable "resource_tags" {
  description = "Tags to set on resource. To specify no tags, set to `null`"
  # Not setting a default, so that this must explicitly be set to `null` to specify no tags
  type     = map(string)
  nullable = true
}
