# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_appmesh_route" "test" {
  name                = var.rName
  mesh_name           = aws_appmesh_mesh.test.id
  virtual_router_name = aws_appmesh_virtual_router.test.name

  spec {
    http_route {
      match {
        prefix = "/"
      }

      action {
        weighted_target {
          virtual_node = aws_appmesh_virtual_node.test1.name
          weight       = 100
        }
      }
    }
  }

  tags = var.resource_tags
}

resource "aws_appmesh_mesh" "test" {
  name = var.rName
}

resource "aws_appmesh_virtual_router" "test" {
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
}

resource "aws_appmesh_virtual_node" "test1" {
  name      = "${var.rName}-1"
  mesh_name = aws_appmesh_mesh.test.id

  spec {
    listener {
      port_mapping {
        port     = 8080
        protocol ="http"
      }
    }

    service_discovery {
      dns {
        hostname = "test1.simpleapp.local"
      }
    }
  }
}

resource "aws_appmesh_virtual_node" "test2" {
  name      = "${var.rName}-2"
  mesh_name = aws_appmesh_mesh.test.id

  spec {
    listener {
      port_mapping {
        port     = 8080
        protocol = "http"
      }
    }

    service_discovery {
      dns {
        hostname = "test2.simpleapp.local"
      }
    }
  }
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
