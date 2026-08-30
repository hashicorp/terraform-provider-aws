resource "aws_bedrockagentcore_workload_identity" "test" {
  name = var.rName

{{- template "tags" . }}
}

