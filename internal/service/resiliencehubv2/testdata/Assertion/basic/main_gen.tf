# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

data "aws_region" "current" {
}

resource "aws_resiliencehubv2_policy" "test" {
  name = "${var.rName}-policy"

  availability_slo {
    target = 99.9
  }
}

resource "aws_resiliencehubv2_service" "test" {
  name    = "${var.rName}-service"
  regions = [data.aws_region.current.name]

  policy_arn = aws_resiliencehubv2_policy.test.arn

  permission_model {
    invoker_role_name = "AWSResilienceHubAssessmentRole"
  }
}

resource "aws_resiliencehubv2_assertion" "test" {
  service_arn = aws_resiliencehubv2_service.test.arn
  text        = "The service must recover within 5 minutes"
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
