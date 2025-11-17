resource "aws_bedrockagentcore_memory" "test" {
{{- template "region" }}
  name                  = var.rName
  event_expiry_duration = 7

{{- template "tags" . }}
}
