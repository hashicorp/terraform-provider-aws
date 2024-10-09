resource "aws_amplify_branch" "test" {
  app_id      = aws_amplify_app.test.id
  branch_name = var.rName
{{- template "tags" . }}
}

resource "aws_amplify_app" "test" {
  name = var.rName
}
