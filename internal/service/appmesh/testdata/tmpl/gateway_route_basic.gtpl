resource "aws_appmesh_gateway_route" "test" {
  name                 = var.rName
  mesh_name            = aws_appmesh_mesh.test.name
  virtual_gateway_name = aws_appmesh_virtual_gateway.test.name

  spec {
    http_route {
      action {
        target {
          virtual_service {
            virtual_service_name = aws_appmesh_virtual_service.test[0].name
          }
        }
      }

      match {
        prefix = "/"
      }
    }
  }

{{- template "tags" . }}
}

resource "aws_appmesh_mesh" "test" {
  name = var.rName
}

resource "aws_appmesh_virtual_gateway" "test" {
  name      = var.rName
  mesh_name = aws_appmesh_mesh.test.name

  spec {
    listener {
      port_mapping {
        port     = 8080
        protocol = "http"
      }
    }
  }
}

resource "aws_appmesh_virtual_service" "test" {
  count = 2

  name      = "${var.rName}-${count.index}"
  mesh_name = aws_appmesh_mesh.test.name

  spec {}
}
