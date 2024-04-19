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

resource "aws_iam_openid_connect_provider" "test" {
  url = "https://accounts.testle.com/${var.rName}"

  client_id_list = [
    "266362248691-re108qaeld573ia0l6clj2i5ac7r7291.apps.testleusercontent.com",
  ]

  thumbprint_list = ["cf23df2207d99a74fbe169e3eba035e633b65d94"]

  tags = {
    (var.tagKey1) = var.tagValue1
    (var.tagKey2) = var.tagValue2
  }
}


variable "rName" {
  type     = string
  nullable = false
}

variable "tagKey1" {
  type     = string
  nullable = false
}

variable "tagValue1" {
  type     = string
  nullable = false
}

variable "tagKey2" {
  type     = string
  nullable = false
}

variable "tagValue2" {
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
