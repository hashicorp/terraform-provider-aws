# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

provider "aws" {
  default_tags {
    tags = var.provider_tags
  }
  ignore_tags {
    keys = var.ignore_tag_keys
  }
}

resource "aws_fms_policy" "test" {
  name                        = var.rName
  delete_all_policy_resources = false
  exclude_resource_tags       = false
  remediation_enabled         = false
  resource_type               = "AWS::EC2::SecurityGroup"

  security_service_policy_data {
    type = "SECURITY_GROUPS_CONTENT_AUDIT"

    managed_service_data = jsonencode({
      type = "SECURITY_GROUPS_CONTENT_AUDIT",

      securityGroupAction = {
        type = "ALLOW"
      },

      securityGroups = [
        {
          id = aws_security_group.test.id
        }
      ],
    })
  }

  tags = var.resource_tags

  depends_on = [aws_fms_admin_account.test]
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_security_group" "test" {
  name   = var.rName
  vpc_id = aws_vpc.test.id

  ingress {
    protocol    = "6"
    from_port   = 80
    to_port     = 8000
    cidr_blocks = ["10.0.0.0/8"]
    description = ""
  }

  egress {
    protocol    = "tcp"
    from_port   = 80
    to_port     = 8000
    cidr_blocks = ["10.0.0.0/8"]
    description = ""
  }
}

# testAccAdminAccountConfig_basic

data "aws_caller_identity" "current" {}

resource "aws_fms_admin_account" "test" {
  account_id = data.aws_caller_identity.current.account_id
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "resource_tags" {
  description = "Tags to set on resource. To specify no tags, set to `null`"
  # Not setting a default, so that this must explicitly be set to `null` to specify no tags
  type     = map(string)
  nullable = true
}

variable "provider_tags" {
  type     = map(string)
  nullable = true
  default  = null
}

variable "ignore_tag_keys" {
  type     = set(string)
  nullable = false
}
