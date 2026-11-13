resource "aws_bedrockagentcore_api_key_credential_provider" "test" {
{{- template "region" }}
  name    = var.rName
  api_key = "test-api-key-value"

{{- template "tags" . }}
}
