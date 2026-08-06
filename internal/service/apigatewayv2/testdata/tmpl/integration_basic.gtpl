resource "aws_apigatewayv2_integration" "test" {
{{- template "region" }}
  api_id           = aws_apigatewayv2_api.test.id
  integration_type = "MOCK"
}

resource "aws_apigatewayv2_api" "test" {
{{- template "region" }}
  name                       = var.rName
  protocol_type              = "WEBSOCKET"
  route_selection_expression = "$request.body.action"
}
