resource "aws_redshift_hsm_client_certificate" "test" {
{{- template "region" }}
  hsm_client_certificate_identifier = var.rName
{{- template "tags" . }}
}
