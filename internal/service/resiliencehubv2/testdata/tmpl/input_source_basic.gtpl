resource "aws_resiliencehubv2_input_source" "test" {
{{- template "region" }}
  service_arn = aws_resiliencehubv2_service.test.arn

  resource_configuration {
    cfn_stack_arn = aws_cloudformation_stack.test.id
  }
}

resource "aws_cloudformation_stack" "test" {
{{- template "region" }}
  name = var.rName

  template_body = jsonencode({
    AWSTemplateFormatVersion = "2010-09-09"
    Description              = "Test stack for NGRH input source"
    Resources = {
      WaitHandle = {
        Type = "AWS::CloudFormation::WaitConditionHandle"
      }
    }
  })
}

resource "aws_resiliencehubv2_service" "test" {
{{- template "region" }}
  name    = var.rName
  regions = [data.aws_region.current.name]

  permission_model {
    invoker_role_name = aws_iam_role.test.name
  }

  depends_on = [aws_iam_role_policy_attachment.service_AWSResilienceHubV2AssessmentExecutionPolicy]
}

data "aws_region" "current" {
{{- template "region" }}
}

data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = "${var.rName}-invoker"

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "resiliencehub.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy_attachment" "service_AWSResilienceHubV2AssessmentExecutionPolicy" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AWSResilienceHubV2AssessmentExecutionPolicy"
  role       = aws_iam_role.test.name
}
