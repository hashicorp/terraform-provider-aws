# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

provider "null" {}

resource "aws_cloudwatchomni_alert" "test" {
  space_id   = aws_cloudwatchomni_space.test.space_id
  profile_id = aws_cloudwatchomni_access_profile.test.profile_id
  name       = var.rName

  rule {
    telemetry_rule {
      query {
        expression = "up"
        language   = "PROMQL"
      }

      condition {
        comparator         = "GT"
        threshold_mode     = "COUNT_OF_RESULTS"
        critical_threshold = 0
      }

      evaluation {
        interval_seconds = 60
      }

      no_data {
        treat_as = "OK"
      }
    }
  }

  tags = {
    (var.unknownTagKey) = null_resource.test.id
    (var.knownTagKey)   = var.knownTagValue
  }

  depends_on = [aws_cloudwatchomni_access_grant.test]
}

# Lets alerts assume the access profile.
resource "aws_cloudwatchomni_access_grant" "test" {
  name       = var.rName
  domain_id  = aws_cloudwatchomni_domain.test.domain_id
  space_id   = aws_cloudwatchomni_space.test.space_id
  permission = "CUSTOM"

  principal {
    principal_type = "ALERT"
    principal_id   = "ALL"
  }

  scoped_action {
    actions = ["cloudwatch:AssumeAccessProfile"]

    resource {
      resource_type = "AccessProfile"
      resource_arns = [aws_cloudwatchomni_access_profile.test.arn]
    }
  }
}

resource "aws_cloudwatchomni_access_profile" "test" {
  name     = var.rName
  space_id = aws_cloudwatchomni_space.test.space_id
}

resource "aws_cloudwatchomni_space" "test" {
  name                 = var.rName
  domain_id            = aws_cloudwatchomni_domain.test.domain_id
  data_access_role_arn = aws_iam_role.test.arn

  depends_on = [aws_iam_role_policy_attachment.test]
}

resource "aws_cloudwatchomni_domain" "test" {
  name               = var.rName
  identity_providers = ["IAM"]
}

resource "aws_iam_role" "test" {
  name = var.rName

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action    = "sts:AssumeRole"
      Effect    = "Allow"
      Principal = { Service = "cloudwatch.${data.aws_partition.current.dns_suffix}" }
      Condition = { StringEquals = { "aws:SourceAccount" = data.aws_caller_identity.current.account_id } }
    }]
  })
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/CloudWatchOmniSpaceAccessPolicy"
}

data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

resource "null_resource" "test" {}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "unknownTagKey" {
  type     = string
  nullable = false
}

variable "knownTagKey" {
  type     = string
  nullable = false
}

variable "knownTagValue" {
  type     = string
  nullable = false
}
