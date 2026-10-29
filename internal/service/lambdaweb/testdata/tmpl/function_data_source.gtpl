data "aws_lambdaweb_function" "test" {
{{- template "region" }}
  function_name = aws_lambdaweb_function.test.function_name
}
