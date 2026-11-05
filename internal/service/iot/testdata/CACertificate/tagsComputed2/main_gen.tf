# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

provider "null" {}

resource "aws_iot_ca_certificate" "test" {
  active                  = true
  allow_auto_registration = true
  ca_certificate_pem      = var.certificate_pem
  certificate_mode        = "SNI_ONLY"

  tags = {
    (var.unknownTagKey) = null_resource.test.id
    (var.knownTagKey)   = var.knownTagValue
  }
}

resource "null_resource" "test" {}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "certificate_pem" {
  type     = string
  nullable = false
}

variable "private_key_pem" {
  type     = string
  nullable = false
}

variable "unknownTagKey" {
  type     = string
  nullable = false
}

variable "knownTagKey" {
  type     = string
  nullable = false
}

variable "knownTagValue" {
  type     = string
  nullable = false
}
