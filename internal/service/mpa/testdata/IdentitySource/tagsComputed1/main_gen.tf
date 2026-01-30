# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

provider "null" {}

data "aws_ssoadmin_instances" "test" {}
data "aws_region" "current" {}

resource "aws_mpa_identity_source" "test" {
  name = var.rName

  identity_source_parameters {
    iam_identity_center {
      instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
      region       = data.aws_region.current.name
    }
  }

  tags = {
    (var.unknownTagKey) = null_resource.test.id
  }
}

resource "null_resource" "test" {}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "unknownTagKey" {
  type     = string
  nullable = false
}
