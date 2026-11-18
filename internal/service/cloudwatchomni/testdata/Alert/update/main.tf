# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_cloudwatchomni_alert" "test" {
  space_id              = aws_cloudwatchomni_space.test.space_id
  profile_id            = var.use_second_profile ? aws_cloudwatchomni_access_profile.test2.profile_id : aws_cloudwatchomni_access_profile.test.profile_id
  name                  = var.alert_name
  description           = var.description
  notifications_enabled = var.notifications_enabled

  rule {
    telemetry_rule {
      query {
        expression = var.expression
        language   = "PROMQL"
      }

      evaluation {
        interval_seconds          = var.interval_seconds
        pending_duration_seconds  = var.pending_duration_seconds
        recovery_duration_seconds = var.recovery_duration_seconds
      }

      condition {
        comparator         = "GT"
        threshold_mode     = "COUNT_OF_RESULTS"
        critical_threshold = var.critical_threshold
        warning_threshold  = var.warning_threshold
      }

      no_data {
        treat_as = var.no_data_treat_as
      }
    }
  }

  dynamic "notification_rule" {
    for_each = var.notification_states

    content {
      target {
        arn  = aws_sns_topic.test.arn
        type = "sns"
      }

      trigger {
        state_values = notification_rule.value
      }
    }
  }

  depends_on = [aws_cloudwatchomni_access_grant.test, aws_cloudwatchomni_access_grant.test2]
}

resource "aws_sns_topic" "test" {
  name = var.rName
}

resource "aws_sns_topic_policy" "test" {
  arn = aws_sns_topic.test.arn

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action    = "sns:Publish"
      Effect    = "Allow"
      Principal = { Service = "cloudwatch.${data.aws_partition.current.dns_suffix}" }
      Resource  = aws_sns_topic.test.arn
      Condition = { StringEquals = { "aws:SourceAccount" = data.aws_caller_identity.current.account_id } }
    }]
  })
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

resource "aws_cloudwatchomni_access_grant" "test2" {
  name       = "${var.rName}-2"
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
      resource_arns = [aws_cloudwatchomni_access_profile.test2.arn]
    }
  }
}

resource "aws_cloudwatchomni_access_profile" "test2" {
  name     = "${var.rName}-2"
  space_id = aws_cloudwatchomni_space.test.space_id
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

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "alert_name" {
  description = "Name of the alert"
  type        = string
  nullable    = false
}

variable "description" {
  description = "Description of the alert"
  type        = string
  default     = null
}

variable "notifications_enabled" {
  description = "Whether notifications are enabled"
  type        = bool
  default     = true
}

variable "expression" {
  description = "PromQL expression to evaluate"
  type        = string
  default     = "up"
}

variable "interval_seconds" {
  description = "Evaluation interval in seconds"
  type        = number
  default     = 60
}

variable "pending_duration_seconds" {
  description = "Pending duration in seconds"
  type        = number
  default     = null
}

variable "recovery_duration_seconds" {
  description = "Recovery duration in seconds"
  type        = number
  default     = null
}

variable "critical_threshold" {
  description = "Critical threshold"
  type        = number
  default     = 0
}

variable "warning_threshold" {
  description = "Warning threshold"
  type        = number
  default     = null
}

variable "no_data_treat_as" {
  description = "State to report when there is no data"
  type        = string
  default     = "OK"
}

variable "notification_states" {
  description = "One notification rule per entry, each triggering on the listed states"
  type        = list(list(string))
  default     = []
}

variable "use_second_profile" {
  description = "Whether the alert uses the second access profile"
  type        = bool
  default     = false
}
