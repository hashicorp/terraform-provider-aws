resource "aws_acm_certificate" "test" {
  certificate_body = var.certificate_pem
  private_key      = var.private_key_pem
{{- template "tags" . }}
}
