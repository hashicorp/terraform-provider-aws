# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_ecs_service" "test" {
  provider = aws

  config {
    # We need to use the cluster name here, because that is the default returned
    # when no cluster is set on the Express Service
    cluster = data.aws_ecs_cluster.default.cluster_name
  }
}
