resource "aws_bedrockagentcore_payment_manager" "test" {
{{- template "region" }}
  name            = var.rName
  authorizer_type = "AWS_IAM"
  role_arn        = aws_iam_role.test.arn

{{- template "tags" . }}
}

data "aws_iam_policy_document" "test" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["bedrock-agentcore.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "test" {
  name               = var.rName
  assume_role_policy = data.aws_iam_policy_document.test.json
}

resource "aws_iam_role_policy" "test" {
  role = aws_iam_role.test.name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect   = "Allow"
        Action   = ["bedrock-agentcore:CreateWorkloadIdentity"]
        Resource = "*"
      }
    ]
  })
}
