resource "aws_bedrockagentcore_policy_engine" "test" {
{{- template "region" }}
  name = var.rName

{{- template "tags" . }}
}
