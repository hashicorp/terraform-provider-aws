# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_ram_principal_association" "test" {
  region = var.region

  principal          = data.aws_caller_identity.receiver.account_id
  resource_share_arn = aws_ram_resource_share.test.id
}

resource "aws_ram_resource_share" "test" {
  region = var.region

  allow_external_principals = true
  name                      = var.rName
}

data "aws_caller_identity" "receiver" {
  provider = "awsalternate"
}

provider "awsalternate" {
  access_key = var.AWS_ALTERNATE_ACCESS_KEY_ID
  profile    = var.AWS_ALTERNATE_PROFILE
  secret_key = var.AWS_ALTERNATE_SECRET_ACCESS_KEY
}

variable "AWS_ALTERNATE_ACCESS_KEY_ID" {
  type     = string
  nullable = true
  default  = null
}

variable "AWS_ALTERNATE_PROFILE" {
  type     = string
  nullable = true
  default  = null
}

variable "AWS_ALTERNATE_SECRET_ACCESS_KEY" {
  type     = string
  nullable = true
  default  = null
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
