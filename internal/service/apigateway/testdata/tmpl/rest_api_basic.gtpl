resource "aws_api_gateway_rest_api" "test" {
{{- template "region" }}
  name = var.rName

{{- template "tags" . }}
}
