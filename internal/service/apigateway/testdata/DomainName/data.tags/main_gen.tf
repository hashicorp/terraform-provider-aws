# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

data "aws_api_gateway_domain_name" "test" {
  domain_name = aws_api_gateway_domain_name.test.domain_name
}

resource "aws_api_gateway_domain_name" "test" {
  domain_name              = var.rName
  regional_certificate_arn = aws_acm_certificate.test.arn

  endpoint_configuration {
    types = ["REGIONAL"]
  }

  tags = var.resource_tags
}

resource "aws_acm_certificate" "test" {
  certificate_body = var.certificate_pem
  private_key      = var.private_key_pem
}

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

variable "resource_tags" {
  description = "Tags to set on resource. To specify no tags, set to `null`"
  # Not setting a default, so that this must explicitly be set to `null` to specify no tags
  type     = map(string)
  nullable = true
}
