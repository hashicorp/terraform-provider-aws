resource "aws_backup_plan" "test" {
  name = var.rName

  rule {
    rule_name         = var.rName
    target_vault_name = aws_backup_vault.test.name
    schedule          = "cron(0 12 * * ? *)"
  }

{{- template "tags" . }}
}

resource "aws_backup_vault" "test" {
  name = var.rName
}
