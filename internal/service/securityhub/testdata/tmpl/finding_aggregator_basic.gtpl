resource "aws_securityhub_finding_aggregator" "test" {
{{- template "region" }}
  linking_mode = "ALL_REGIONS"

  depends_on = [aws_securityhub_account.test]
}

resource "aws_securityhub_account" "test" {
{{- template "region" }}
}
