# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_ecs_daemon" "test" {
  provider = aws

  config {
    cluster_arn = aws_ecs_cluster.test.arn
  }

  include_resource = true
}
