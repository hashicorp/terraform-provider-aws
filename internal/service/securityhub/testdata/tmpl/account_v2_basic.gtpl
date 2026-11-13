resource "aws_securityhub_account_v2" "test" {
{{- template "region" }}
{{- template "tags" . }}
}