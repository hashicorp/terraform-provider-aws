resource "aws_api_gateway_resource" "test" {
{{- template "region" }}
  rest_api_id = aws_api_gateway_rest_api.test.id
  parent_id   = aws_api_gateway_rest_api.test.root_resource_id
  path_part   = "test"
}

resource "aws_api_gateway_rest_api" "test" {
{{- template "region" }}
  name = var.rName
}
