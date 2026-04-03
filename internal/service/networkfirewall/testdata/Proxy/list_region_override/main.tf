# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

data "aws_availability_zones" "available" {
  region = var.region
  state  = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  region = var.region

  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = var.rName
  }
}

resource "aws_subnet" "public" {
  region = var.region

  vpc_id                  = aws_vpc.test.id
  cidr_block              = "10.0.1.0/24"
  availability_zone       = data.aws_availability_zones.available.names[0]
  map_public_ip_on_launch = true

  tags = {
    Name = "${var.rName}-public"
  }
}

resource "aws_internet_gateway" "test" {
  region = var.region

  vpc_id = aws_vpc.test.id

  tags = {
    Name = var.rName
  }
}

resource "aws_eip" "test" {
  count  = var.resource_count
  region = var.region

  domain = "vpc"

  tags = {
    Name = "${var.rName}-${count.index}"
  }

  depends_on = [aws_internet_gateway.test]
}

resource "aws_nat_gateway" "test" {
  count  = var.resource_count
  region = var.region

  allocation_id = aws_eip.test[count.index].id
  subnet_id     = aws_subnet.public.id

  tags = {
    Name = "${var.rName}-${count.index}"
  }

  depends_on = [aws_internet_gateway.test]
}

resource "aws_networkfirewall_proxy_configuration" "test" {
  region = var.region

  name = var.rName

  default_rule_phase_actions {
    post_response = "ALLOW"
    pre_dns       = "ALLOW"
    pre_request   = "ALLOW"
  }
}

resource "aws_networkfirewall_proxy" "test" {
  count  = var.resource_count
  region = var.region

  name                    = "${var.rName}-${count.index}"
  nat_gateway_id          = aws_nat_gateway.test[count.index].id
  proxy_configuration_arn = aws_networkfirewall_proxy_configuration.test.arn

  tls_intercept_properties {
    tls_intercept_mode = "DISABLED"
  }

  listener_properties {
    port = 8080
    type = "HTTP"
  }

  listener_properties {
    port = 443
    type = "HTTPS"
  }
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "resource_count" {
  description = "Number of resources to create"
  type        = number
  nullable    = false
}

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
