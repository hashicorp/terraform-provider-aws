# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_resiliencehubv2_policy" "test" {
  region = var.region

  name = "${var.rName}-policy"

  availability_slo {
    target = 99.9
  }
}

resource "aws_resiliencehubv2_service" "test" {
  region = var.region

  name       = "${var.rName}-service"
  regions    = [var.region]
  policy_arn = aws_resiliencehubv2_policy.test.arn

  permission_model {
    invoker_role_name = "AWSResilienceHubAssessmentRole"
  }
}

resource "aws_cloudformation_stack" "test" {
  count  = var.resource_count
  region = var.region

  name = "${var.rName}-${count.index}"

  template_body = jsonencode({
    AWSTemplateFormatVersion = "2010-09-09"
    Description              = "Test stack for NGRH input source"
    Resources = {
      WaitHandle = {
        Type = "AWS::CloudFormation::WaitConditionHandle"
      }
    }
  })
}

resource "aws_resiliencehubv2_input_source" "test" {
  count  = var.resource_count
  region = var.region

  service_arn   = aws_resiliencehubv2_service.test.arn
  cfn_stack_arn = aws_cloudformation_stack.test[count.index].id
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "resource_count" {
  description = "Number of resources to create"
  type        = number
  nullable    = false
}

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
