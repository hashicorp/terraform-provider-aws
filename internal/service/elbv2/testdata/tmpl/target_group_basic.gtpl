resource "aws_lb_target_group" "test" {
{{- template "region" }}
  name     = var.rName
  port     = 443
  protocol = "HTTPS"
  vpc_id   = aws_vpc.test.id

{{- template "tags" . }}
}

resource "aws_vpc" "test" {
{{- template "region" }}
  cidr_block = "10.0.0.0/16"
}
