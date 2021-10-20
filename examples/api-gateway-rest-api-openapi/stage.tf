#
# Stage and Stage Settings
#

resource "aws_apigateway_stage" "example" {
  deployment_id = aws_apigateway_deployment.example.id
  rest_api_id   = aws_apigateway_rest_api.example.id
  stage_name    = "example"
}

resource "aws_apigateway_method_settings" "example" {
  rest_api_id = aws_apigateway_rest_api.example.id
  stage_name  = aws_apigateway_stage.example.stage_name
  method_path = "*/*"

  settings {
    metrics_enabled = true
  }
}
