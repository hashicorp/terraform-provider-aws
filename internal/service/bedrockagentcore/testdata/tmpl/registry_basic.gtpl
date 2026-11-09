resource "aws_bedrockagentcore_registry" "test" {
{{- template "region" }}
  name = var.rName
}
