resource "aws_backup_vault" "test" {
{{- template "region" }}
  name = var.rName

{{- template "tags" . }}
}
