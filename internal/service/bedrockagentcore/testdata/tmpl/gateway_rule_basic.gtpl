resource "aws_bedrockagentcore_gateway_rule" "test" {
{{- template "region" }}
  gateway_identifier = aws_bedrockagentcore_gateway.test.gateway_id
  priority           = 100

  action {
    route_to_target {
      static_route {
        target_name = aws_bedrockagentcore_gateway_target.test.name
      }
    }
  }
}

resource "aws_bedrockagentcore_gateway" "test" {
{{- template "region" }}
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
{{- template "region" }}
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

resource "aws_iam_role" "gateway" {
  name               = "${var.rName}-gateway"
  assume_role_policy = data.aws_iam_policy_document.gateway_assume.json
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

resource "aws_bedrockagentcore_agent_runtime" "test" {
{{- template "region" }}
  agent_runtime_name = replace(var.rName, "-", "_")
  role_arn           = aws_iam_role.runtime.arn

  agent_runtime_artifact {
    code_configuration {
      entry_point = ["main.py"]
      runtime     = "PYTHON_3_13"
      code {
        s3 {
          bucket = aws_s3_bucket.test.bucket
          prefix = aws_s3_object.test.key
        }
      }
    }
  }

  network_configuration {
    network_mode = "PUBLIC"
  }

  protocol_configuration {
    server_protocol = "HTTP"
  }

  depends_on = [aws_iam_role_policy.bucket]
}

resource "aws_iam_role" "runtime" {
  name               = "${var.rName}-runtime"
  assume_role_policy = data.aws_iam_policy_document.runtime_assume.json
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

# testAccAgentRuntimeConfig_baseS3Bucket

resource "aws_s3_bucket" "test" {
{{- template "region" }}
  bucket        = replace(var.rName, "_", "-")
  force_destroy = true
}

resource "aws_s3_object" "test" {
{{- template "region" . }}
  bucket = aws_s3_bucket.test.bucket
  key    = "agent-runtime-codezip.zip"
  source = "${path.module}/test-fixtures/agent-runtime-codezip.zip"
}

resource "aws_iam_role_policy" "bucket" {
  name   = var.rName
  role   = aws_iam_role.runtime.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect   = "Allow"
      Action   = ["s3:GetObject", "s3:GetObjectVersion"]
      Resource = "${aws_s3_bucket.test.arn}/*"
    }]
  })
}
