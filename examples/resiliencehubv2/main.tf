# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

terraform {
  required_version = ">= 0.12"
}

provider "aws" {
  region = "us-west-2"
}

data "aws_region" "current" {}

# A reusable resilience policy defining availability SLO and DR targets.
resource "aws_resiliencehubv2_policy" "example" {
  name = "terraform-example-policy"

  availability_slo {
    target = 99.9
  }

  multi_az {
    disaster_recovery_approach = "ACTIVE_ACTIVE"
    rpo_in_minutes             = 5
    rto_in_minutes             = 10
  }
}

# A top-level system that groups services together.
resource "aws_resiliencehubv2_system" "example" {
  name        = "terraform-example-system"
  description = "Example Resilience Hub system"
}

# A service assessed against the policy above.
resource "aws_resiliencehubv2_service" "example" {
  name    = "terraform-example-service"
  regions = [data.aws_region.current.name]

  policy_arn = aws_resiliencehubv2_policy.example.arn

  permission_model {
    invoker_role_name = "AWSResilienceHubAssessmentRole"
  }
}

# A critical user journey within the system.
resource "aws_resiliencehubv2_user_journey" "example" {
  name       = "terraform-example-journey"
  system_arn = aws_resiliencehubv2_system.example.arn
}

# A technical workflow subset of the service.
resource "aws_resiliencehubv2_service_function" "example" {
  name        = "terraform-example-function"
  service_arn = aws_resiliencehubv2_service.example.arn
  criticality = "PRIMARY"
}

# A CloudFormation stack used as a resource-discovery input source.
resource "aws_cloudformation_stack" "example" {
  name = "terraform-example-stack"

  template_body = jsonencode({
    AWSTemplateFormatVersion = "2010-09-09"
    Description              = "Example stack for Resilience Hub input source"
    Resources = {
      WaitHandle = {
        Type = "AWS::CloudFormation::WaitConditionHandle"
      }
    }
  })
}

# An input source that discovers resources from the CloudFormation stack.
resource "aws_resiliencehubv2_input_source" "example" {
  service_arn   = aws_resiliencehubv2_service.example.arn
  cfn_stack_arn = aws_cloudformation_stack.example.id
}
