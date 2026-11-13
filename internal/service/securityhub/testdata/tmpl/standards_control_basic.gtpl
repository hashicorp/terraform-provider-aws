resource "aws_securityhub_standards_control" "test" {
{{- template "region" }}
  standards_control_arn = format("%s/1.10", replace(aws_securityhub_standards_subscription.test.id, "subscription", "control"))
  control_status        = "ENABLED"
}

resource "aws_securityhub_standards_subscription" "test" {
{{- template "region" }}
  standards_arn = "arn:${data.aws_partition.current.partition}:securityhub:::ruleset/cis-aws-foundations-benchmark/v/1.2.0"
  depends_on    = [aws_securityhub_account.test]
}

data "aws_partition" "current" {}

resource "aws_securityhub_account" "test" {
{{- template "region" }}
}
