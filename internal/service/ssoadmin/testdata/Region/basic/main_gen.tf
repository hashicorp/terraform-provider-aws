# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_region" "test" {
  instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
  region_name  = "us-west-2"
}

