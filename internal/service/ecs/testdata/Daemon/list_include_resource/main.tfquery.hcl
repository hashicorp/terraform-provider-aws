# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

list "aws_ecs_daemon" "test" {
  provider = aws

  config {
    cluster = aws_ecs_cluster.test.arn
  }

  include_resource = true
}
