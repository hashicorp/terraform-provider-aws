resource "aws_detective_organization_configuration" "test" {
{{- template "region" }}
  auto_enable = true
  graph_arn   = aws_detective_graph.test.graph_arn

  depends_on = [aws_detective_organization_admin_account.test]
}

resource "aws_detective_graph" "test" {
{{- template "region" }}
}

resource "aws_detective_organization_admin_account" "test" {
{{- template "region" }}
  account_id = data.aws_caller_identity.current.account_id
}

data "aws_caller_identity" "current" {}