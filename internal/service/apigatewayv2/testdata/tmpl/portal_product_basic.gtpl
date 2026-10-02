resource "aws_apigatewayv2_portal_product" "test" {
{{- template "region" }}
  display_name = var.rName

{{- template "tags" . }}
}
