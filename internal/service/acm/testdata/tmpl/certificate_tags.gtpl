resource "aws_acm_certificate" "test" {
{{- template "region" }}
  certificate_body = var.certificate_pem
  private_key      = var.private_key_pem
{{- template "tags" . }}
}
