resource "aws_bedrockagentcore_policy_engine" "test" {
{{- template "region" }}
  name = replace(var.rName, "-", "_")

{{- template "tags" . }}
}
