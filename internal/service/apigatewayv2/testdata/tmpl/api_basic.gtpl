resource "aws_apigatewayv2_api" "test" {
{{- template "region" }}
  name                       = var.rName
  protocol_type              = "WEBSOCKET"
  route_selection_expression = "$request.body.action"

{{- template "tags" . }}
}
