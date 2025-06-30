# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_ssoadmin_application_assignment_configuration" "test" {
  region = var.region

  application_arn     = aws_ssoadmin_application.test.application_arn
  assignment_required = true
}

resource "aws_ssoadmin_application" "test" {
  region = var.region

  name                     = var.rName
  application_provider_arn = local.test_application_provider_arn
  instance_arn             = tolist(data.aws_ssoadmin_instances.test.arns)[0]
}

data "aws_ssoadmin_instances" "test" {
  region = var.region
}

locals {
  test_application_provider_arn = "arn:aws:sso::aws:applicationProvider/custom"
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
