# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_ecs_express_gateway_service" "test" {
  execution_role_arn      = aws_iam_role.execution.arn
  infrastructure_role_arn = aws_iam_role.infrastructure.arn

  primary_container {
    image = "public.ecr.aws/nginx/nginx:1.28-alpine3.21-slim"
  }

  depends_on = [
    aws_iam_role_policy_attachment.execution,
    aws_iam_role_policy_attachment.infrastructure,
  ]
}

# Used in `query.tfquery.hcl`
# tflint-ignore: terraform_unused_declarations
data "aws_ecs_cluster" "default" {
  cluster_name = "default"

  depends_on = [aws_ecs_express_gateway_service.test]
}

# testAccExpressGatewayServiceConfig_base

data "aws_partition" "current" {}

resource "aws_iam_role" "execution" {
  name               = "${var.rName}-execution"
  assume_role_policy = <<POLICY
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "ecs-tasks.amazonaws.com"
        ]
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy_attachment" "execution" {
  role       = aws_iam_role.execution.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

resource "aws_iam_role_policy" "execution_logs" {
  name = "CreateLogGroupPolicy"
  role = aws_iam_role.execution.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect   = "Allow"
      Action   = "logs:CreateLogGroup"
      Resource = "*"
    }]
  })
}

resource "aws_iam_role" "infrastructure" {
  name               = "${var.rName}-infra"
  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "ecs.amazonaws.com"
        ]
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy_attachment" "infrastructure" {
  role       = aws_iam_role.infrastructure.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonECSInfrastructureRoleforExpressGatewayServices"
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
