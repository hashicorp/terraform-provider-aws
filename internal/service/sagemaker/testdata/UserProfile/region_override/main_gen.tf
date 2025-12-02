# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_sagemaker_user_profile" "test" {
  region = var.region

  domain_id         = aws_sagemaker_domain.test.id
  user_profile_name = var.rName
}

# testAccUserProfileConfig_base

resource "aws_sagemaker_domain" "test" {
  region = var.region

  domain_name = var.rName
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = aws_subnet.test[*].id

  default_user_settings {
    execution_role = aws_iam_role.test.arn
  }

  retention_policy {
    home_efs_file_system = "Delete"
  }
}

resource "aws_iam_role" "test" {
  name               = var.rName
  path               = "/"
  assume_role_policy = data.aws_iam_policy_document.test.json
}

data "aws_iam_policy_document" "test" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["sagemaker.amazonaws.com"]
    }
  }
}

# acctest.ConfigVPCWithSubnets(rName, 1)

resource "aws_vpc" "test" {
  region = var.region

  cidr_block = "10.0.0.0/16"
}

# acctest.ConfigSubnets(rName, 1)

resource "aws_subnet" "test" {
  region = var.region

  count = 1

  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
}

# acctest.ConfigAvailableAZsNoOptInDefaultExclude

data "aws_availability_zones" "available" {
  region = var.region

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
