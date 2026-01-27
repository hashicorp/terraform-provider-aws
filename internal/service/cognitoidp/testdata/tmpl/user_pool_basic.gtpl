resource "aws_cognito_user_pool" "test" {
  name = var.rName
{{- template "tags" . }}
}
