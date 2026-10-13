resource "aws_mailmanager_archive" "test" {
{{- template "region" }}
  name = var.rName
{{- template "tags" . }}
}
