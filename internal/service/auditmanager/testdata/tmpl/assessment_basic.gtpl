resource "aws_auditmanager_assessment" "test" {
{{- template "region" }}
  name = var.rName

  assessment_reports_destination {
    destination      = "s3://${aws_s3_bucket.test.bucket}"
    destination_type = "S3"
  }

  framework_id = aws_auditmanager_framework.test.id

  roles {
    role_arn  = aws_iam_role.test.arn
    role_type = "PROCESS_OWNER"
  }

  scope {
    aws_accounts {
      id = data.aws_caller_identity.current.account_id
    }
  }

{{ template "tags" . }}
}

resource "aws_s3_bucket" "test" {
{{- template "region" }}
  bucket        = var.rName
  force_destroy = true
}

resource "aws_iam_role" "test" {
  name = var.rName

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Sid    = ""
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      },
    ]
  })
}

resource "aws_auditmanager_control" "test" {
{{- template "region" }}
  name = var.rName

  control_mapping_sources {
    source_name          = var.rName
    source_set_up_option = "Procedural_Controls_Mapping"
    source_type          = "MANUAL"
  }
}

resource "aws_auditmanager_framework" "test" {
{{- template "region" }}
  name = var.rName

  control_sets {
    name = var.rName

    controls {
      id = aws_auditmanager_control.test.id
    }
  }
}

data "aws_caller_identity" "current" {}
