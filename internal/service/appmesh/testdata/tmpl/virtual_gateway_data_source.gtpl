data "aws_appmesh_virtual_gateway" "test" {
  name      = aws_appmesh_virtual_gateway.test.name
  mesh_name = aws_appmesh_mesh.test.name
}
