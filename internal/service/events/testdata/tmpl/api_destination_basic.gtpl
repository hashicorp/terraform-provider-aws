resource "aws_cloudwatch_event_api_destination" "test" {
{{- template "region" }}
  name                = var.rName
  invocation_endpoint = "https://example.com/"
  http_method         = "GET"
  connection_arn      = aws_cloudwatch_event_connection.test.arn
}

resource "aws_cloudwatch_event_connection" "test" {
{{- template "region" }}
  name               = "${var.rName}-conn"
  authorization_type = "API_KEY"
  auth_parameters {
    api_key {
      key   = "testKey"
      value = "testValue"
    }
  }
}
