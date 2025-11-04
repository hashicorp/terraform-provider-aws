# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

provider "null" {}

resource "aws_dms_endpoint" "test" {
  database_name = "tf-test-dms-db"
  endpoint_id   = var.rName
  endpoint_type = "source"
  engine_name   = "aurora"
  password      = "tftest"
  port          = 3306
  server_name   = "tftest"
  ssl_mode      = "none"
  username      = "tftest"

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
