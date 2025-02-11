data "aws_appmesh_virtual_node" "test" {
  name      = aws_appmesh_virtual_node.test.name
  mesh_name = aws_appmesh_mesh.test.name
}
