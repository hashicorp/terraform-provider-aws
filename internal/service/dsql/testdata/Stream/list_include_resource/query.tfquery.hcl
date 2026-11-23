# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_dsql_stream" "test" {
  provider = aws

  include_resource = true

  config {
    cluster_identifier = aws_dsql_cluster.test.identifier
  }
}
