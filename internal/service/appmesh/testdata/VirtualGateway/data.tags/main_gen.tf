# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

data "aws_appmesh_virtual_gateway" "test" {
  name      = aws_appmesh_virtual_gateway.test.name
  mesh_name = aws_appmesh_mesh.test.name
}

resource "aws_appmesh_virtual_gateway" "test" {
  name      = var.rName
  mesh_name = aws_appmesh_mesh.test.id

  spec {
    listener {
      port_mapping {
        port     = 8080
        protocol = "http"
      }
    }
  }

  tags = var.resource_tags
}

resource "aws_appmesh_mesh" "test" {
  name = var.rName
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
