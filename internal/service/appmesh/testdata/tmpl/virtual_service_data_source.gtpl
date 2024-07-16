data "aws_appmesh_virtual_service" "test" {
  name      = aws_appmesh_virtual_service.test.name
  mesh_name = aws_appmesh_mesh.test.name
}
