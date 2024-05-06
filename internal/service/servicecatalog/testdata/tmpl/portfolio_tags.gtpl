resource "aws_servicecatalog_portfolio" "test" {
  name          = var.rName
  description   = "test-b"
  provider_name = "test-c"
{{- template "tags" . }}
}
