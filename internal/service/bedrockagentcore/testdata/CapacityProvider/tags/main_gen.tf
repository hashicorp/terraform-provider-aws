# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0


resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}
resource "aws_subnet" "test" {
  vpc_id     = aws_vpc.test.id
  cidr_block = "10.0.1.0/24"
}
resource "aws_security_group" "test" {
  name   = var.rName
  vpc_id = aws_vpc.test.id
}
resource "aws_iam_role" "capacity_provider" {
  name = var.rName
  assume_role_policy = jsonencode({
    Version   = "2012-10-17"
    Statement = [{ Effect = "Allow", Action = "sts:AssumeRole", Principal = { Service = "bedrock-agentcore.amazonaws.com" } }]
  })
}
data "aws_partition" "current" {}
resource "aws_iam_role_policy_attachment" "capacity_provider" {
  role       = aws_iam_role.capacity_provider.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/BedrockAgentCoreRuntimeInstancesOperatorRolePolicy"
}
resource "aws_bedrockagentcore_capacity_provider" "test" {
  depends_on = [aws_iam_role_policy_attachment.capacity_provider]

  tags = var.resource_tags
  name = var.rName
  compute_configuration {
    ec2_configuration {
      launch_template_source {
        launch_parameters {
          operating_system = "LINUX_ARM64"
          instance_requirements {
            allowed_instance_types = ["c6g.medium"]
          }
        }
      }
      vpc_configuration {
        subnets         = [aws_subnet.test.id]
        security_groups = [aws_security_group.test.id]
      }
    }
  }
  permissions_configuration {
    capacity_provider_operator_role_arn = aws_iam_role.capacity_provider.arn
  }
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "resource_tags" {
  description = "Tags to set on resource. To specify no tags, set to `null`"
  # Not setting a default, so that this must explicitly be set to `null` to specify no tags
  type     = map(string)
  nullable = true
}
