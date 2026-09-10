resource "aws_cloudwatch_log_delivery_destination" "test" {
{{- template "region" }}
  name = var.rName

  delivery_destination_configuration {
    destination_resource_arn = aws_cloudwatch_log_group.test.arn
  }
}

resource "aws_cloudwatch_log_group" "test" {
{{- template "region" }}
  name = var.rName
}
