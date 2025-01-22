# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

provider "null" {}

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

  tags = {
    (var.unknownTagKey) = null_resource.test.id
  }
}

resource "aws_appmesh_mesh" "test" {
  name = var.rName
}

resource "aws_appmesh_virtual_node" "test" {
  name      = var.rName
  mesh_name = aws_appmesh_mesh.test.id

  spec {}
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
