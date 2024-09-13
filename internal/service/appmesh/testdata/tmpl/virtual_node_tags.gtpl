resource "aws_appmesh_virtual_node" "test" {
  name      = var.rName
  mesh_name = aws_appmesh_mesh.test.id

  spec {}

{{- template "tags" . }}
}

resource "aws_appmesh_mesh" "test" {
  name =var.rName
}
