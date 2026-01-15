resource "aws_rds_certificate" "test" {
{{- template "region" }}
  certificate_identifier = "rds-ca-rsa4096-g1"
}
