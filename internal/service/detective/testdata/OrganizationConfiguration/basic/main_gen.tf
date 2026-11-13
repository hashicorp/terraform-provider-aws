# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_detective_organization_configuration" "test" {
  auto_enable = true
  graph_arn   = aws_detective_graph.test.graph_arn

  depends_on = [aws_detective_organization_admin_account.test]
}

resource "aws_detective_graph" "test" {
}

resource "aws_detective_organization_admin_account" "test" {
  account_id = data.aws_caller_identity.current.account_id
}

data "aws_caller_identity" "current" {}
