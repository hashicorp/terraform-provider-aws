resource "aws_iot_ca_certificate" "test" {
{{- template "region" }}
  active                  = true
  allow_auto_registration = true
  ca_certificate_pem      = var.certificate_pem
  certificate_mode        = "SNI_ONLY"
{{- template "tags" . }}
}
