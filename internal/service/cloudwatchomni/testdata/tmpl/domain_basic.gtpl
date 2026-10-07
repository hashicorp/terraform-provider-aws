resource "aws_cloudwatchomni_domain" "test" {
{{- template "region" }}
  name               = var.rName
  identity_providers = ["IAM"]
}
