resource "aws_cloudfront_connection_function" "test" {
  name                     = var.rName
  connection_function_code = "function handler(event) { return event.request; }"

  connection_function_config {
    comment = "Test connection function"
    runtime = "cloudfront-js-2.0"
  }

{{- template "tags" . }}
}
