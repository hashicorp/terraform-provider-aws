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

{{- template "tags" . }}
}

resource "aws_appmesh_mesh" "test" {
  name = var.rName
}

resource "aws_appmesh_virtual_node" "test" {
  name      = var.rName
  mesh_name = aws_appmesh_mesh.test.id

  spec {}
}
