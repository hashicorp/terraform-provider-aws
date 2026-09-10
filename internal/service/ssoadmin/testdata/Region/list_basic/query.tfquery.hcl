# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_ssoadmin_region" "test" {
  provider = aws

  config {
    instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
  }
}
