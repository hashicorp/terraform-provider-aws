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

{{- template "tags" . }}
}

resource "aws_appmesh_mesh" "test" {
  name = var.rName
}
