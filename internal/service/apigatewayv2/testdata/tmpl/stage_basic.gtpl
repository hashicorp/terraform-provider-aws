resource "aws_apigatewayv2_stage" "test" {
  name   = var.rName
  api_id = aws_apigatewayv2_api.test.id

{{- template "tags" . }}
}

resource "aws_apigatewayv2_api" "test" {
  name                       = var.rName
  protocol_type              = "WEBSOCKET"
  route_selection_expression = "$request.body.action"
}
