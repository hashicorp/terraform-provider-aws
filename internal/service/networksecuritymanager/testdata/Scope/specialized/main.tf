# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_networksecuritymanager_scope" "test" {
  name                   = var.rName
  specialized_scope_type = var.specialized_scope_type

  scope_configuration {
    resource_scope {
      resource_type = "AWS::EC2::Subnet"
      include_all   = true
    }
  }
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "specialized_scope_type" {
  type     = string
  nullable = true
  default  = null
}
