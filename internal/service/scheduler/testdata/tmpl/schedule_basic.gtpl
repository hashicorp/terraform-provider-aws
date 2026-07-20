resource "aws_scheduler_schedule" "test" {
{{- template "region" }}
  name = var.rName

  flexible_time_window {
    mode = "OFF"
  }

  schedule_expression = "rate(1 hour)"

  target {
    arn      = aws_sqs_queue.test.arn
    role_arn = aws_iam_role.test.arn
  }
}

resource "aws_sqs_queue" "test" {
{{- template "region" }}
}

resource "aws_iam_role" "test" {
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = {
      Effect = "Allow"
      Action = "sts:AssumeRole"
      Principal = {
        Service = "scheduler.${data.aws_partition.main.dns_suffix}"
      }
      Condition = {
        StringEquals = {
          "aws:SourceAccount" : data.aws_caller_identity.main.account_id
        }
      }
    }
  })
}

data "aws_caller_identity" "main" {}

data "aws_partition" "main" {}
