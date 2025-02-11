data "aws_appmesh_route" "test" {
  name                = aws_appmesh_route.test.name
  mesh_name           = aws_appmesh_route.test.mesh_name
  virtual_router_name = aws_appmesh_route.test.virtual_router_name
}
