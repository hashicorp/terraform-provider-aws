# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

provider "null" {}

data "aws_ssoadmin_instances" "test" {}

resource "aws_accountaccess_application" "test" {
  identity_center_instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]

  tags = {
    (var.unknownTagKey) = null_resource.test.id
  }

  lifecycle {
    precondition {
      condition     = length(var.rName) > 0
      error_message = "The generated test name must not be empty."
    }
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
