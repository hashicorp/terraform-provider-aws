resource "aws_globalaccelerator_accelerator" "test" {
  {{- template "region" . }}
  name            = var.rName
  ip_address_type = "IPV4"
  enabled         = false

  {{- template "tags" . }}
}
