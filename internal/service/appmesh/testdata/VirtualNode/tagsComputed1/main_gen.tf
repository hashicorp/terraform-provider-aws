# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

provider "null" {}

resource "aws_appmesh_virtual_node" "test" {
  name      = var.rName
  mesh_name = aws_appmesh_mesh.test.id

  spec {}

  tags = {
    (var.unknownTagKey) = null_resource.test.id
  }
}

resource "aws_appmesh_mesh" "test" {
  name =var.rName
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
