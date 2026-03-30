resource "aws_vpc_security_group_ingress_rule" "test" {
{{- template "region" }}
  security_group_id = aws_security_group.test.id

  cidr_ipv4   = "10.0.0.0/8"
  from_port   = 80
  ip_protocol = "tcp"
  to_port     = 8080

{{- template "tags" . }}
}

resource "aws_vpc" "test" {
{{- template "region" }}
  cidr_block = "10.0.0.0/16"
}

resource "aws_security_group" "test" {
{{- template "region" }}
  vpc_id = aws_vpc.test.id
  name   = var.rName
}
