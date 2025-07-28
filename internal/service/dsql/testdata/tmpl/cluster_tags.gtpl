resource "aws_dsql_cluster" "test" {
  deletion_protection_enabled = false

{{- template "tags" . }}
}

output "rName" {
  value       = var.rName
  description = "To prevent tflint issues"
}
