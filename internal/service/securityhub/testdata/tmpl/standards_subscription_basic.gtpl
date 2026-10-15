resource "aws_securityhub_standards_subscription" "test" {
{{- template "region" }}
  standards_arn = "arn:${data.aws_partition.current.partition}:securityhub:::ruleset/cis-aws-foundations-benchmark/v/1.2.0"
  depends_on    = [aws_securityhub_account.test]
}

resource "aws_securityhub_account" "test" {
{{- template "region" }}
}

data "aws_partition" "current" {}