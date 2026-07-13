resource "aws_bedrockagentcore_online_evaluation_config" "test" {
{{- template "region" }}
  online_evaluation_config_name = var.rName
  enable_on_create              = false
  evaluation_execution_role_arn = aws_iam_role.test.arn

  data_source_config {
    cloudwatch_logs {
      log_group_names = [aws_cloudwatch_log_group.test.name]
      service_names   = ["strands_healthcare_single_agent.DEFAULT"]
    }
  }

  evaluator {
    evaluator_id = "Builtin.Helpfulness"
  }

  rule {
    sampling_config {
      sampling_percentage = 10.0
    }
  }

{{- template "tags" . }}
}

data "aws_partition" "current" {}
data "aws_region" "current" {}
data "aws_caller_identity" "current" {}

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

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "CloudWatchLogReadStatement",
      "Effect": "Allow",
      "Action": [
        "logs:DescribeLogGroups",
        "logs:GetQueryResults",
        "logs:StartQuery"
      ],
      "Resource": "*"
    },
    {
      "Sid": "CloudWatchLogWriteStatement",
      "Effect": "Allow",
      "Action": [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource": "arn:${data.aws_partition.current.partition}:logs:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}:log-group:/aws/bedrock-agentcore/evaluations/*"
    },
    {
      "Sid": "CloudWatchIndexPolicyStatement",
      "Effect": "Allow",
      "Action": [
        "logs:DescribeIndexPolicies",
        "logs:PutIndexPolicy"
      ],
      "Resource": [
        "arn:${data.aws_partition.current.partition}:logs:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}:log-group:aws/spans",
        "arn:${data.aws_partition.current.partition}:logs:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}:log-group:aws/spans:*"
      ]
    },
    {
      "Sid": "BedrockInvokeStatement",
      "Effect": "Allow",
      "Action": [
        "bedrock:InvokeModel",
        "bedrock:InvokeModelWithResponseStream"
      ],
      "Resource": [
        "arn:${data.aws_partition.current.partition}:bedrock:${data.aws_region.current.region}::foundation-model/*",
        "arn:${data.aws_partition.current.partition}:bedrock:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}:inference-profile/*"
      ]
    }
  ]
}
EOF
}

resource "aws_cloudwatch_log_group" "test" {
{{- template "region" }}
  name = "/aws/agentcore/${var.rName}"
}
