# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

provider "null" {}

resource "aws_ssmcontacts_contact" "test" {
  alias = var.rName
  type  = "PERSONAL"

  tags = {
    (var.unknownTagKey) = null_resource.test.id
  }

  depends_on = [aws_ssmincidents_replication_set.test]
}

# testAccContactConfig_base

resource "aws_ssmincidents_replication_set" "test" {
  region {
    name = data.aws_region.current.name
  }
}

data "aws_region" "current" {}

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
