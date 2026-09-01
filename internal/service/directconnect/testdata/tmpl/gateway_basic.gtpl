resource "aws_dx_gateway" "test" {
  name            = var.rName
  amazon_side_asn = var.rBgpAsn
{{- template "tags" . }}
}
