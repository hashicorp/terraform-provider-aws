resource "aws_globalaccelerator_accelerator" "example" {
{{- template "region" }}
  name            = var.rName
  ip_address_type = "IPV4"
  enabled         = false
}

resource "aws_globalaccelerator_listener" "test" {
{{- template "region" }}
  accelerator_arn = aws_globalaccelerator_accelerator.example.arn
  protocol        = "TCP"

  port_range {
    from_port = 80
    to_port   = 81
  }
}
