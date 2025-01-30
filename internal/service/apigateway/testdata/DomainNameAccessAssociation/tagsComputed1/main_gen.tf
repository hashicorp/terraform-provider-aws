# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

provider "null" {}

data "aws_region" "current" {}

resource "aws_vpc" "test" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = {
    Name = var.rName
  }
}

resource "aws_default_security_group" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, 0)
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = var.rName
  }
}

resource "aws_vpc_endpoint" "test" {
  private_dns_enabled = false
  security_group_ids  = [aws_default_security_group.test.id]
  service_name        = "com.amazonaws.${data.aws_region.current.name}.execute-api"
  subnet_ids          = [aws_subnet.test.id]
  vpc_endpoint_type   = "Interface"
  vpc_id              = aws_vpc.test.id

  tags = {
    Name = var.rName
  }
}

resource "aws_acm_certificate" "test" {
  certificate_body = var.certificate_pem
  private_key      = var.private_key_pem
}

resource "aws_api_gateway_domain_name" "test" {
  domain_name     = var.rName
  certificate_arn = aws_acm_certificate.test.arn

  endpoint_configuration {
    types = ["PRIVATE"]
  }
}

resource "aws_api_gateway_domain_name_access_association" "test" {
  access_association_source      = aws_vpc_endpoint.test.id
  access_association_source_type = "VPCE"
  domain_name_arn                = aws_api_gateway_domain_name.test.arn

  tags = {
    (var.unknownTagKey) = null_resource.test.id
  }
}

# acctest.ConfigAvailableAZsNoOptInDefaultExclude()

data "aws_availability_zones" "available" {
  exclude_zone_ids = local.default_exclude_zone_ids
  state            = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

locals {
  default_exclude_zone_ids = ["usw2-az4", "usgw1-az2"]
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
