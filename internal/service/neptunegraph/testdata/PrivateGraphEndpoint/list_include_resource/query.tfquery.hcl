# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_neptunegraph_private_graph_endpoint" "test" {
  provider = aws

  include_resource = true

  config {
    graph_identifier = aws_neptunegraph_graph.test.id
  }
}
