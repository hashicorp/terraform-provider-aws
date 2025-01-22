# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_appmesh_virtual_service" "test" {
  name      = var.rName
  mesh_name = aws_appmesh_mesh.test.id

  spec {
    provider {
      virtual_node {
        virtual_node_name = aws_appmesh_virtual_node.test.name
      }
    }
  }

  tags = var.resource_tags
}

resource "aws_appmesh_mesh" "test" {
  name = var.rName
}

resource "aws_appmesh_virtual_node" "test" {
  name      = var.rName
  mesh_name = aws_appmesh_mesh.test.id

  spec {}
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
