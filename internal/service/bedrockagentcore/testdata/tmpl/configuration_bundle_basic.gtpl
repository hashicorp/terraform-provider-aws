resource "aws_bedrockagentcore_configuration_bundle" "test" {
{{- template "region" }}
  bundle_name = var.rName

  component {
    component_identifier = "arn:aws:bedrock-agentcore:::evaluator/Builtin.Helpfulness"
    configuration        = jsonencode({ threshold = 0.7 })
  }
{{- template "tags" . }}
}
