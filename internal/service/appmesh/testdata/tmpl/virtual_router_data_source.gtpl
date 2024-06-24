data "aws_appmesh_virtual_router" "test" {
  name      = aws_appmesh_virtual_router.test.name
  mesh_name = aws_appmesh_mesh.test.name
}
