# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_ssm_maintenance_window_task" "test" {
  window_id        = aws_ssm_maintenance_window.test.id
  task_type        = "RUN_COMMAND"
  task_arn         = "AWS-RunShellScript"
  priority         = 1
  service_role_arn = aws_iam_role.test.arn
  max_concurrency  = "2"
  max_errors       = "1"

  targets {
    key    = "WindowTargetIds"
    values = [aws_ssm_maintenance_window_target.test.id]
  }

  task_invocation_parameters {
    run_command_parameters {
      parameter {
        name   = "commands"
        values = ["pwd"]
      }
    }
  }
}

resource "aws_ssm_maintenance_window" "test" {
  cutoff   = 1
  duration = 3
  name     = var.rName
  schedule = "cron(0 16 ? * TUE *)"
}

resource "aws_ssm_maintenance_window_target" "test" {
  name          = var.rName
  resource_type = "INSTANCE"
  window_id     = aws_ssm_maintenance_window.test.id

  targets {
    key    = "tag:Name"
    values = ["tf-acc-test"]
  }
}

resource "aws_iam_role" "test" {
  name = var.rName

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Principal = {
          Service = "events.amazonaws.com"
        }
        Effect = "Allow"
        Sid    = ""
      }
    ]
  })
}

resource "aws_iam_role_policy" "test" {
  name = var.rName
  role = aws_iam_role.test.name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = {
      Effect   = "Allow"
      Action   = "ssm:*"
      Resource = "*"
    }
  })
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
