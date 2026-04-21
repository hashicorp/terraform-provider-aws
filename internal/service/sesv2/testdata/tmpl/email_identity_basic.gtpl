resource "aws_sesv2_email_identity" "test" {
  email_identity = var.rName
{{- template "tags" . }}
}
