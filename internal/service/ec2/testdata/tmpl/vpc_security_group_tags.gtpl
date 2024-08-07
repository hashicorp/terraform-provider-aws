resource "aws_security_group" "test" {
  name   = var.rName
  vpc_id = aws_vpc.test.id

{{- template "tags" . }}
}

resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"
}
