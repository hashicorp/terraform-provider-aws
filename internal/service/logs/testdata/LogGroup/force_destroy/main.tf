# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_cloudwatch_log_group" "test" {
  name          = var.rName
  force_destroy = true

  tags = var.resource_tags
}
