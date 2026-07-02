# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_ssoadmin_region" "test" {
  instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
  region_name  = "us-east-1"
}

data "aws_ssoadmin_instances" "test" {}

