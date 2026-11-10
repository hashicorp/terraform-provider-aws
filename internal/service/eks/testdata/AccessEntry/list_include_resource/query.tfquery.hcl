# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_eks_access_entry" "test" {
  provider = aws

  include_resource = true

  config {
    cluster_name = aws_eks_cluster.test.name
  }
}
