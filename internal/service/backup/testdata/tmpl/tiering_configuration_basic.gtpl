resource "aws_backup_tiering_configuration" "test" {
{{- template "region" }}
  name              = replace(var.rName, "-", "_")
  backup_vault_name = aws_backup_vault.test.name

  resource_selection {
    resource_type                 = "S3"
    resources                     = ["*"]
    tiering_down_settings_in_days = 60
  }

{{- template "tags" . }}
}

resource "aws_backup_vault" "test" {
{{- template "region" }}
  name = var.rName
}
