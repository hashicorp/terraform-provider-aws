resource "aws_arcregionswitch_plan" "test" {
{{- template "region" }}
  name              = var.rName
  execution_role    = aws_iam_role.test.arn
  recovery_approach = "activePassive"
  regions           = [local.primary_region, local.secondary_region]
  primary_region    = local.primary_region

  workflow {
    workflow_target_action = "activate"
    workflow_target_region = local.secondary_region
    step {
      name                 = "basic-step"
      execution_block_type = "ManualApproval"

      execution_approval_config {
        approval_role   = aws_iam_role.test.arn
        timeout_minutes = 60
      }
    }
  }

  workflow {
    workflow_target_action = "activate"
    workflow_target_region = local.primary_region

    step {
      name                 = "basic-step-primary"
      execution_block_type = "ManualApproval"

      execution_approval_config {
        approval_role   = aws_iam_role.test.arn
        timeout_minutes = 60
      }
    }
  }

{{- template "tags" . }}
}

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

locals {
  primary_region   = data.aws_region.current.name
  secondary_region = data.aws_region.secondary.name
}

data "aws_region" "current" {
{{- template "region" -}}
}

data "aws_region" "secondary" {
  region = var.secondary_region
}
