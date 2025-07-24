resource "aws_globalaccelerator_accelerator" "test" {
{{- template "region" }}
  name = var.rName
}

resource "aws_globalaccelerator_listener" "test" {
{{- template "region" }}
  accelerator_arn = aws_globalaccelerator_accelerator.test.arn

  port_range {
    from_port = 80
    to_port   = 80
  }
}

resource "aws_globalaccelerator_endpoint_group" "test" {
{{- template "region" }}
  listener_arn = aws_globalaccelerator_listener.test.arn
{{- template "tags" }}
}
