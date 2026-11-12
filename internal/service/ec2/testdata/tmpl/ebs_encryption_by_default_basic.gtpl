resource "aws_ebs_encryption_by_default" "test" {
{{- template "region" }}
  enabled = true
}
