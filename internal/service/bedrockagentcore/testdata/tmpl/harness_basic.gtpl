resource "aws_bedrockagentcore_harness" "test" {
{{- template "region" }}
  harness_name       = var.rName
  execution_role_arn = aws_iam_role.test.arn

  model {
    bedrock_model_config {
      model_id = "anthropic.claude-sonnet-4-20250514"
    }
  }

  system_prompt {
    text = "You are a helpful assistant."
  }

{{- template "tags" . }}
}

resource "aws_iam_role" "test" {
  name               = var.rName
  assume_role_policy = data.aws_iam_policy_document.test.json
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

resource "aws_iam_role_policy" "test" {
  role = aws_iam_role.test.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": [
      "bedrock:InvokeModel",
      "bedrock:InvokeModelWithResponseStream"
    ],
    "Resource": "*"
  }
}
EOF
}
