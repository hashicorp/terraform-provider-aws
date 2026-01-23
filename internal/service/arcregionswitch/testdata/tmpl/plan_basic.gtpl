resource "aws_iam_role" "test" {
  name = var.rName

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "arc-region-switch.amazonaws.com"
        }
      },
    ]
  })
}

resource "aws_arcregionswitch_plan" "test" {
{{- template "region" }}
  name              = var.rName
  execution_role    = aws_iam_role.test.arn
  recovery_approach = "activePassive"
  regions           = [data.aws_region.alternate.region, data.aws_region.current.region]
  primary_region    = data.aws_region.alternate.region

{{- template "tags" . }}

  workflow {
    workflow_target_action = "activate"
    workflow_target_region = data.aws_region.alternate.region

    step {
      name                 = "minimal-step-secondary"
      execution_block_type = "ManualApproval"

      execution_approval_config {
        approval_role   = aws_iam_role.test.arn
        timeout_minutes = 60
      }
    }
  }

  workflow {
    workflow_target_action = "activate"
    workflow_target_region = data.aws_region.current.region

    step {
      name                 = "minimal-step-primary"
      execution_block_type = "ManualApproval"

      execution_approval_config {
        approval_role   = aws_iam_role.test.arn
        timeout_minutes = 60
      }
    }
  }
}

data "aws_region" "current" {}

data "aws_region" "alternate" {
  provider = "awsalternate"
}

{{ template "acctest.ConfigAlternateAccountProvider" }}
