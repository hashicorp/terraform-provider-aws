resource "aws_redshift_hsm_configuration" "test" {
{{- template "region" }}
  description                   = var.rName
  hsm_configuration_identifier  = var.rName
  hsm_ip_address                = "10.0.0.1"
  hsm_partition_name            = "aws"
  hsm_partition_password        = var.rName
  hsm_server_public_certificate = var.rName
{{- template "tags" . }}
}
