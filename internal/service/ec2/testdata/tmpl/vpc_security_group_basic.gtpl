resource "aws_security_group" "test" {
{{- template "region" }}
  name   = var.rName
  vpc_id = aws_vpc.test.id

{{- template "tags" . }}
}

resource "aws_vpc" "test" {
{{- template "region" }}
  cidr_block = "10.1.0.0/16"
}
