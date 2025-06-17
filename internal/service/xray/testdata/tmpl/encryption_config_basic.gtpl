resource "aws_xray_encryption_config" "test" {
{{- template "region" }}
  type = "NONE"
}
