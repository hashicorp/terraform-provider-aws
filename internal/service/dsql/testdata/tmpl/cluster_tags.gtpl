resource "aws_dsql_cluster" "test" {
{{- template "tags" . }}
}

output "rName" {
  value       = var.rName
  description = "To prevent tflint issues"
}
