resource "aws_bedrockagentcore_browser_profile" "test" {
{{- template "region" }}
  name = var.rName

{{- template "tags" . }}
}