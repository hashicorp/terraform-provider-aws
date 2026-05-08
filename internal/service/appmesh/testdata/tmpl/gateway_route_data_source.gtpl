data "aws_appmesh_gateway_route" "test" {
  name                 = aws_appmesh_gateway_route.test.name
  mesh_name            = aws_appmesh_gateway_route.test.mesh_name
  virtual_gateway_name = aws_appmesh_gateway_route.test.virtual_gateway_name
}
