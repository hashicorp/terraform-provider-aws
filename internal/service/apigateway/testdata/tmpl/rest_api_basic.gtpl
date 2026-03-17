resource "aws_api_gateway_rest_api" "test" {
  name = var.rName

{{- template "tags" . }}
}
