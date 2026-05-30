resource "aws_apigatewayv2_route" "test" {
{{- template "region" }}
  api_id    = aws_apigatewayv2_api.test.id
  route_key = "$default"
}

resource "aws_apigatewayv2_api" "test" {
{{- template "region" }}
  name                       = var.rName
  protocol_type              = "WEBSOCKET"
  route_selection_expression = "$request.body.action"
}
