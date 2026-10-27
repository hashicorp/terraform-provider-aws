# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

data "aws_outposts_outposts" "test" {
}

data "aws_outposts_outpost_instance_types" "test" {
  arn = tolist(data.aws_outposts_outposts.test.arns)[0]
}

resource "aws_outposts_capacity_task" "test" {
  outpost_identifier = tolist(data.aws_outposts_outposts.test.arns)[0]

  instance_pool {
    instance_type = tolist(data.aws_outposts_outpost_instance_types.test.instance_types)[0]
    count         = 1
  }
}

