resource "aws_api_gateway_api_key" "test" {
  name = var.rName

{{- template "tags" . }}
}
