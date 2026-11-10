resource "aws_bedrock_guardrail_resource_policy" "test" {
{{- template "region" }}
  resource_arn = aws_bedrock_guardrail.test.guardrail_arn

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = { AWS = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root" }
      Action    = ["bedrock:ApplyGuardrail", "bedrock:GetGuardrail"]
      Resource  = aws_bedrock_guardrail.test.guardrail_arn
    }]
  })
}

data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

resource "aws_bedrock_guardrail" "test" {
{{- template "region" }}
  name                      = var.rName
  blocked_input_messaging   = "test"
  blocked_outputs_messaging = "test"
  description               = "test"

  content_policy_config {
    filters_config {
      input_strength  = "MEDIUM"
      output_strength = "MEDIUM"
      type            = "HATE"
    }
  }
}
