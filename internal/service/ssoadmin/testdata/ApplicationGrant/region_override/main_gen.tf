# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

data "aws_ssoadmin_instances" "test" {
  region = var.region
}

resource "aws_ssoadmin_application" "test" {
  region = var.region

  name                     = var.rName
  application_provider_arn = "arn:aws:sso::aws:applicationProvider/custom"
  instance_arn             = tolist(data.aws_ssoadmin_instances.test.arns)[0]
}

resource "aws_ssoadmin_application_grant" "test" {
  region = var.region

  application_arn = aws_ssoadmin_application.test.application_arn
  grant_type      = "authorization_code"

  grant {
    authorization_code {
      redirect_uris = ["https://example.com/callback"]
    }
  }
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
