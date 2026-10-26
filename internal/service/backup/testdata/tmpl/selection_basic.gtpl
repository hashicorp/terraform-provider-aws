resource "aws_backup_selection" "test" {
{{- template "region" }}
  plan_id = aws_backup_plan.test.id

  name         = var.rName
  iam_role_arn = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:role/service-role/AWSBackupDefaultServiceRole"

  selection_tag {
    type  = "STRINGEQUALS"
    key   = "foo"
    value = "bar"
  }

  resources = [
    "arn:${data.aws_partition.current.partition}:ec2:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:volume/*"
  ]
}

resource "aws_backup_plan" "test" {
{{- template "region" }}
  name = var.rName

  rule {
    rule_name         = var.rName
    target_vault_name = aws_backup_vault.test.name
    schedule          = "cron(0 12 * * ? *)"
  }
}

resource "aws_backup_vault" "test" {
{{- template "region" }}
  name = var.rName
}

data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

data "aws_region" "current" {
{{- template "region" -}}
}
