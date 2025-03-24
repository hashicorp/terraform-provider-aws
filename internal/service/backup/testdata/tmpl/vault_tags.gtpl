resource "aws_backup_vault" "test" {
  name = var.rName

{{- template "tags" . }}
}
