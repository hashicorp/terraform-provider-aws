resource "aws_bedrock_enforced_guardrail_configuration" "test" {
{{- template "region" }}
  guardrail_arn     = aws_bedrock_guardrail.test.guardrail_arn
  guardrail_version = aws_bedrock_guardrail_version.test.version
}

resource "aws_bedrock_guardrail" "test" {
{{- template "region" }}
  name                      = var.rName
  blocked_input_messaging   = "Blocked input"
  blocked_outputs_messaging = "Blocked output"
  description               = "Test guardrail for enforced guardrail configuration"

  word_policy_config {
    words_config {
      text = "deny"
    }
  }
}

resource "aws_bedrock_guardrail_version" "test" {
{{- template "region" }}
  guardrail_arn = aws_bedrock_guardrail.test.guardrail_arn
  description   = "Test guardrail version"
}