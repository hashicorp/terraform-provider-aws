# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

provider "aws" {
  default_tags {
    tags = var.provider_tags
  }
}

# tflint-ignore: terraform_unused_declarations
data "aws_accountaccess_application" "test" {
  arn = aws_accountaccess_application.test.arn
}

data "aws_ssoadmin_instances" "test" {
}

resource "aws_accountaccess_application" "test" {
  identity_source {
    identity_center {
      instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
    }
  }

  tags = var.resource_tags
}

variable "resource_tags" {
  description = "Tags to set on resource. To specify no tags, set to `null`"
  # Not setting a default, so that this must explicitly be set to `null` to specify no tags
  type     = map(string)
  nullable = true
}

variable "provider_tags" {
  type     = map(string)
  nullable = false
}
