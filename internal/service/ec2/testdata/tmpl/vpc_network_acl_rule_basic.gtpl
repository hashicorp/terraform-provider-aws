resource "aws_network_acl_rule" "test" {
{{- template "region" }}
  network_acl_id = aws_network_acl.test.id
  rule_number    = 200
  egress         = true
  protocol       = "tcp"
  rule_action    = "allow"
  cidr_block     = "0.0.0.0/0"
  from_port      = 22
  to_port        = 22
}

resource "aws_vpc" "test" {
{{- template "region" }}
  cidr_block = "10.3.0.0/16"

  tags = {
    Name = var.rName
  }
}

resource "aws_network_acl" "test" {
{{- template "region" }}
  vpc_id = aws_vpc.test.id

  tags = {
    Name = var.rName
  }
}
