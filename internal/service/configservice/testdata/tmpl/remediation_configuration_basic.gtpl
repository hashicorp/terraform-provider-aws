resource "aws_config_remediation_configuration" "test" {
{{- template "region" }}
  config_rule_name = aws_config_config_rule.test.name

  resource_type  = "AWS::S3::Bucket"
  target_id      = "AWS-EnableS3BucketEncryption"
  target_type    = "SSM_DOCUMENT"
  target_version = "1"

  parameter {
    name         = "AutomationAssumeRole"
    static_value = aws_iam_role.test.arn
  }
  parameter {
    name           = "BucketName"
    resource_value = "RESOURCE_ID"
  }
  parameter {
    name         = "SSEAlgorithm"
    static_value = "AES256"
  }
  automatic                  = false
  maximum_automatic_attempts = 2
  retry_attempt_seconds      = 30
  execution_controls {
    ssm_controls {
      concurrent_execution_rate_percentage = 75
      error_percentage                     = 10
    }
  }
}

resource "aws_sns_topic" "test" {
{{- template "region" }}
  name = var.rName
}

resource "aws_config_config_rule" "test" {
{{- template "region" }}
  name = var.rName

  source {
    owner             = "AWS"
    source_identifier = "S3_BUCKET_VERSIONING_ENABLED"
  }

  depends_on = [aws_config_configuration_recorder.test]
}

resource "aws_config_configuration_recorder" "test" {
{{- template "region" }}
  name     = var.rName
  role_arn = aws_iam_role.test.arn
}

resource "aws_iam_role" "test" {
  name = var.rName

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "config.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = var.rName
  role = aws_iam_role.test.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
        "Action": "config:Put*",
        "Effect": "Allow",
        "Resource": "*"

    }
  ]
}
EOF
}