# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

data "aws_outposts_outposts" "test" {
  region = var.region

}

data "aws_outposts_outpost_instance_types" "test" {
  region = var.region

  arn = tolist(data.aws_outposts_outposts.test.arns)[0]
}

resource "aws_outposts_capacity_task" "test" {
  region = var.region

  outpost_identifier = tolist(data.aws_outposts_outposts.test.arns)[0]

  instance_pool {
    instance_type = tolist(data.aws_outposts_outpost_instance_types.test.instance_types)[0]
    count         = 1
  }
}


variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
