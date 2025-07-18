resource "aws_cognito_user_pool" "test" {
  name = var.name
  {{- template "region" }}
}

resource "aws_iam_role" "lambda" {
  name = var.name
  {{- template "region" }}

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      }
    ]
  })
}

data "aws_iam_policy_document" "lambda_basic" {
  statement {
    effect = "Allow"
    actions = [
      "logs:CreateLogGroup",
      "logs:CreateLogStream",
      "logs:PutLogEvents"
    ]
    resources = ["arn:aws:logs:*:*:*"]
  }
}

resource "aws_iam_role_policy" "lambda_basic" {
  name   = var.name
  role   = aws_iam_role.lambda.id
  policy = data.aws_iam_policy_document.lambda_basic.json
  {{- template "region" }}
}

resource "aws_lambda_function" "test" {
  filename         = "test-fixtures/lambdatest.zip"
  function_name    = var.name
  role            = aws_iam_role.lambda.arn
  handler         = "index.handler"
  runtime         = "nodejs18.x"
  source_code_hash = filebase64sha256("test-fixtures/lambdatest.zip")

  depends_on = [aws_iam_role_policy.lambda_basic]
  {{- template "region" }}
}

resource "aws_lambda_permission" "appsync_invoke" {
  statement_id  = "AllowExecutionFromAppSync"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.test.function_name
  principal     = "appsync.amazonaws.com"
  {{- template "region" }}
}

resource "aws_iam_role" "cloudwatch" {
  name = var.name
  {{- template "region" }}

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "appsync.amazonaws.com"
        }
      }
    ]
  })
}

data "aws_iam_policy_document" "cloudwatch_logs" {
  statement {
    effect = "Allow"
    actions = [
      "logs:CreateLogGroup",
      "logs:CreateLogStream",
      "logs:PutLogEvents"
    ]
    resources = ["arn:aws:logs:*:*:*"]
  }
}

resource "aws_iam_role_policy" "cloudwatch" {
  name   = var.name
  role   = aws_iam_role.cloudwatch.id
  policy = data.aws_iam_policy_document.cloudwatch_logs.json
  {{- template "region" }}
}

resource "aws_appsync_event_api" "test" {
  name          = var.name
  owner_contact = "test@example.com"
  {{- template "region" }}

  event_config {
    auth_providers {
      auth_type = "AMAZON_COGNITO_USER_POOLS"
      cognito_config {
        user_pool_id = aws_cognito_user_pool.test.id
        aws_region   = data.aws_region.current.name
      }
    }

    auth_providers {
      auth_type = "AWS_LAMBDA"
      lambda_authorizer_config {
        authorizer_uri                  = aws_lambda_function.test.arn
        authorizer_result_ttl_in_seconds = 300
      }
    }

    auth_providers {
      auth_type = "OPENID_CONNECT"
      openid_connect_config {
        issuer    = "https://example.com"
        client_id = "test-client-id"
      }
    }

    connection_auth_modes {
      auth_type = "AWS_LAMBDA"
    }

    default_publish_auth_modes {
      auth_type = "AWS_LAMBDA"
    }

    default_subscribe_auth_modes {
      auth_type = "AWS_LAMBDA"
    }

    log_config {
      cloudwatch_logs_role_arn = aws_iam_role.cloudwatch.arn
      log_level                = "ERROR"
    }
  }

  depends_on = [
    aws_lambda_permission.appsync_invoke,
    aws_iam_role_policy.cloudwatch
  ]
}

data "aws_region" "current" {}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}