resource "aws_lambda_layer_version" "test" {
{{- template "region" }}
  filename   = "test-fixtures/lambdatest.zip"
  layer_name = var.rName
}
