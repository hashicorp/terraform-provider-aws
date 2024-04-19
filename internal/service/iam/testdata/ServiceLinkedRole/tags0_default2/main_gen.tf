# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

provider "aws" {
  default_tags {
    tags = {
      (var.providerTagKey1) = var.providerTagValue1
      (var.providerTagKey2) = var.providerTagValue2
    }
  }
}

resource "aws_iam_service_linked_role" "test" {
  aws_service_name = "autoscaling.amazonaws.com"
  custom_suffix    = var.rName

}


variable "rName" {
  type     = string
  nullable = false
}


variable "providerTagKey1" {
  type     = string
  nullable = false
}

variable "providerTagValue1" {
  type     = string
  nullable = false
}


variable "providerTagKey2" {
  type     = string
  nullable = false
}

variable "providerTagValue2" {
  type     = string
  nullable = false
}
