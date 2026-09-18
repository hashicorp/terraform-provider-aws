# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_bedrockagentcore_gateway_rule" "test" {
  count = var.resource_count

  gateway_identifier = aws_bedrockagentcore_gateway.test.gateway_id
  priority           = 100 + count.index

  action {
    route_to_target {
      static_route {
        target_name = aws_bedrockagentcore_gateway_target.test.name
      }
    }
  }
}

data "aws_iam_policy_document" "gateway_assume" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["bedrock-agentcore.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "gateway" {
  name               = "${var.rName}-gateway"
  assume_role_policy = data.aws_iam_policy_document.gateway_assume.json
}

resource "aws_bedrockagentcore_gateway" "test" {
  name     = "${var.rName}-gateway"
  role_arn = aws_iam_role.gateway.arn

  authorizer_type = "CUSTOM_JWT"
  authorizer_configuration {
    custom_jwt_authorizer {
      discovery_url    = "https://accounts.google.com/.well-known/openid-configuration"
      allowed_audience = ["test"]
    }
  }
}

resource "aws_bedrockagentcore_gateway_target" "test" {
  name               = "${var.rName}-target"
  gateway_identifier = aws_bedrockagentcore_gateway.test.gateway_id

  credential_provider_configuration {
    gateway_iam_role {}
  }

  target_configuration {
    http {
      agentcore_runtime {
        arn = aws_bedrockagentcore_agent_runtime.test.agent_runtime_arn
      }
    }
  }
}

resource "aws_bedrockagentcore_agent_runtime" "test" {
  agent_runtime_name = var.rNameRuntime
  role_arn           = aws_iam_role.runtime.arn

  agent_runtime_artifact {
    container_configuration {
      container_uri = var.image_uri
    }
  }

  network_configuration {
    network_mode = "PUBLIC"
  }

  protocol_configuration {
    server_protocol = "HTTP"
  }
}

data "aws_iam_policy_document" "runtime_assume" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["bedrock-agentcore.amazonaws.com"]
    }
  }
}

data "aws_iam_policy_document" "runtime" {
  statement {
    actions = [
      "ecr:GetAuthorizationToken",
      "ecr:BatchGetImage",
      "ecr:GetDownloadUrlForLayer"
    ]
    effect    = "Allow"
    resources = ["*"]
  }
}

resource "aws_iam_role" "runtime" {
  name               = "${var.rName}-runtime"
  assume_role_policy = data.aws_iam_policy_document.runtime_assume.json
}

resource "aws_iam_role_policy" "runtime" {
  role   = aws_iam_role.runtime.id
  policy = data.aws_iam_policy_document.runtime.json
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "resource_count" {
  description = "Number of resources to create"
  type        = number
  nullable    = false
}

variable "image_uri" {
  type     = string
  nullable = false
}

variable "rNameRuntime" {
  description = "Name for aws_bedreockagentcore_runtime resource"
  type        = string
  nullable    = false
}