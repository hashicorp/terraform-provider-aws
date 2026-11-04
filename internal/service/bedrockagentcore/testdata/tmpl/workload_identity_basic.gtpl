resource "aws_bedrockagentcore_workload_identity" "test" {
{{- template "region" }}
  name = var.rName

{{- template "tags" . }}
}

