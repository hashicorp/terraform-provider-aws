resource "aws_globalaccelerator_custom_routing_accelerator" "test" {
{{- template "region" }}
  name = var.rName
}

resource "aws_globalaccelerator_custom_routing_listener" "test" {
{{- template "region" }}
  accelerator_arn = aws_globalaccelerator_custom_routing_accelerator.test.arn

  port_range {
    from_port = 443
    to_port   = 443
  }
}

resource "aws_globalaccelerator_custom_routing_endpoint_group" "test" {
{{- template "region" }}
  listener_arn = aws_globalaccelerator_custom_routing_listener.test.arn

  destination_configuration {
    from_port = 443
    to_port   = 8443
    protocols = ["TCP"]
  }
}
