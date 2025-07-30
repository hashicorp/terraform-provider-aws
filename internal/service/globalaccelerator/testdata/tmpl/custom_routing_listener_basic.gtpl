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

  port_range {
    from_port = 10000
    to_port   = 30000
  }
}
