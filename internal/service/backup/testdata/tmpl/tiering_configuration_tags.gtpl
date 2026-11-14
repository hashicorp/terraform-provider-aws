resource "aws_backup_tiering_configuration" "test" {
  backup_vault_name          = aws_backup_vault.test.name
  tiering_configuration_name = var.rName

  resource_selection {
    resource_type                 = "S3"
    resources                     = ["*"]
    tiering_down_settings_in_days = 90
  }

{{- template "tags" . }}
}

resource "aws_backup_vault" "test" {
  name = var.rName
}
