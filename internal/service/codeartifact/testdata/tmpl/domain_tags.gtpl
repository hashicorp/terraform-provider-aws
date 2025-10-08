resource "aws_codeartifact_domain" "test" {
{{- template "region" }}
  domain         = var.rName
  encryption_key = aws_kms_key.test.arn
{{- template "tags" . }}
}

resource "aws_kms_key" "test" {
{{- template "region" }}
  description             = var.rName
  deletion_window_in_days = 7
  enable_key_rotation     = true
}
