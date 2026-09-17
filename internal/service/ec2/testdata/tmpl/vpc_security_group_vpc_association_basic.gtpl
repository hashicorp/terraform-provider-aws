resource "aws_vpc_security_group_vpc_association" "test" {
{{- template "region" }}
  security_group_id = aws_security_group.test.id
  vpc_id            = aws_vpc.target.id
}

resource "aws_vpc" "source" {
{{- template "region" }}
  cidr_block = "10.6.0.0/16"
}

resource "aws_security_group" "test" {
{{- template "region" }}
  name   = var.rName
  vpc_id = aws_vpc.source.id
}

resource "aws_vpc" "target" {
{{- template "region" }}
  cidr_block = "10.7.0.0/16"
}
