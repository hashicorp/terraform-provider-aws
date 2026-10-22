resource "aws_lambdaweb_function" "test" {
{{- template "region" }}
  function_name = var.rName

{{- template "tags" . }}
}
