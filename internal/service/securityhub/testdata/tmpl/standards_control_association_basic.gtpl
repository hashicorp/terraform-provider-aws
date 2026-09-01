resource "aws_securityhub_standards_control_association" "test" {
{{- template "region" }}
  security_control_id = "IAM.1"
  standards_arn       = aws_securityhub_standards_subscription.test.standards_arn
  association_status  = "ENABLED"
}

data "aws_partition" "current" {}

resource "aws_securityhub_account" "test" {
{{- template "region" }}
  enable_default_standards = false
}

resource "aws_securityhub_standards_subscription" "test" {
{{- template "region" }}
  standards_arn = "arn:${data.aws_partition.current.partition}:securityhub:::ruleset/cis-aws-foundations-benchmark/v/1.2.0"
  depends_on    = [aws_securityhub_account.test]
}