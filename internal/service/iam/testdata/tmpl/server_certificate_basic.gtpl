resource "aws_iam_server_certificate" "test" {
  name             = var.rName
  certificate_body = var.certificate_pem
  private_key      = var.private_key_pem
{{- template "tags" . }}
}
