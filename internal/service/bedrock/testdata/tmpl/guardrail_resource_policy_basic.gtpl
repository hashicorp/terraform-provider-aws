resource "aws_bedrock_guardrail_resource_policy" "test" {
{{- template "region" }}
  resource_arn = aws_bedrock_guardrail.test.guardrail_arn

  policy = data.aws_iam_policy_document.test.json
}

data "aws_iam_policy_document" "test" {
  statement {
    effect    = "Allow"
    actions   = ["bedrock:ApplyGuardrail", "bedrock:GetGuardrail"]
    resources = [aws_bedrock_guardrail.test.guardrail_arn]

    principals {
      type        = "AWS"
      identifiers = ["arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"]
    }
  }
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
