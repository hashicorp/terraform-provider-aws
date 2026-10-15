resource "aws_amplify_app" "test" {
  name = var.rName
{{- template "tags" . }}
}
