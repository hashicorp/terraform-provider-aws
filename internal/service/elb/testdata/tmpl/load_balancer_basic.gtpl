resource "aws_elb" "test" {
{{- template "region" }}

  name = var.rName

  internal = true
  subnets  = aws_subnet.test[*].id

  listener {
    instance_port     = 8000
    instance_protocol = "http"
    lb_port           = 80
    lb_protocol       = "http"
  }

  cross_zone_load_balancing = true

{{- template "tags" . }}
}

{{ template "acctest.ConfigVPCWithSubnets" 2 }}
