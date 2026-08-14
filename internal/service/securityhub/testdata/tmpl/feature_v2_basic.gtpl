resource "aws_securityhub_feature_v2" "test" {
{{- template "region" }}
  feature_name = "NETWORK_SCANNING"

  depends_on = [aws_securityhub_account_v2.test]
}

resource "aws_securityhub_account_v2" "test" {
{{- template "region" }}
}
