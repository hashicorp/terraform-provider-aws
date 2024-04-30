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

resource "aws_servicecatalog_portfolio" "test" {
  name          = var.rName
  description   = "test-b"
  provider_name = "test-c"

  tags = var.tags
}


variable "rName" {
  type     = string
  nullable = false
}

variable "tags" {
  type     = map(string)
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
