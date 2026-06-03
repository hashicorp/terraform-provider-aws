resource "aws_api_gateway_method" "test" {
{{- template "region" }}
  rest_api_id   = aws_api_gateway_rest_api.test.id
  resource_id   = aws_api_gateway_resource.test.id
  http_method   = "GET"
  authorization = "NONE"

  request_models = {
    "application/json" = "Error"
  }

  request_parameters = {
    "method.request.header.Content-Type" = false
    "method.request.querystring.page"    = true
  }
}

resource "aws_api_gateway_rest_api" "test" {
{{- template "region" }}
  name = var.rName
}

resource "aws_api_gateway_resource" "test" {
{{- template "region" }}
  rest_api_id = aws_api_gateway_rest_api.test.id
  parent_id   = aws_api_gateway_rest_api.test.root_resource_id
  path_part   = "test"
}
