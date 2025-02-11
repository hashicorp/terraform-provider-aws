resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

{{- template "tags" . }}
}

resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"
}
