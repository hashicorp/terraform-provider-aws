resource "aws_backup_framework" "test" {
  name        = var.rName
  description = var.rName

  control {
    name = "BACKUP_RESOURCES_PROTECTED_BY_BACKUP_PLAN"

    scope {
      compliance_resource_types = [
        "EBS"
      ]
    }
  }

{{- template "tags" . }}
}
