# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

data "aws_region" "current" {
  region = var.region

}

resource "aws_resiliencehubv2_policy" "test" {
  region = var.region

  name = "${var.rName}-policy"

  availability_slo {
    target = 99.9
  }
}

resource "aws_resiliencehubv2_service" "test" {
  region = var.region

  name    = "${var.rName}-service"
  regions = [data.aws_region.current.name]

  policy_arn = aws_resiliencehubv2_policy.test.arn

  permission_model {
    invoker_role_name = "AWSResilienceHubAssessmentRole"
  }
}

resource "aws_resiliencehubv2_service_function" "test" {
  region = var.region

  name        = var.rName
  service_arn = aws_resiliencehubv2_service.test.arn
  criticality = "PRIMARY"
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
