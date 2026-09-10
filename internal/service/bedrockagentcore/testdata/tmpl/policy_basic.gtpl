resource "aws_bedrockagentcore_policy" "test" {
{{- template "region" }}
  name             = var.rName
  policy_engine_id = aws_bedrockagentcore_policy_engine.test.policy_engine_id
  validation_mode  = "IGNORE_ALL_FINDINGS"

  definition {
    cedar {
      statement = "permit(principal, action, resource is AgentCore::Gateway);"
    }
  }
}

resource "aws_bedrockagentcore_policy_engine" "test" {
{{- template "region" }}
  name = var.rName
}