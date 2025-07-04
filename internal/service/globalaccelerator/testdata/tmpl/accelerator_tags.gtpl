resource "aws_globalaccelerator_accelerator" "test" {
  {{- template "region" . }}
  name            = var.rName
  ip_address_type = "IPV4"
  enabled         = false

  tags = var.resource_tags
}

variable "resource_tags" {
  description = "Tags to set on resource. To specify no tags, set to `null`"
  # Not setting a type to allow for `null` value
  default = null
}
