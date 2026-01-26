resource "aws_route" "test" {
{{- template "region" }}
  route_table_id         = aws_route_table.test.id
  destination_cidr_block = "10.3.0.0/16"
  gateway_id             = aws_internet_gateway.test.id
}

resource "aws_vpc" "test" {
{{- template "region" }}
  cidr_block = "10.1.0.0/16"
}

resource "aws_internet_gateway" "test" {
{{- template "region" }}
  vpc_id = aws_vpc.test.id
}

resource "aws_route_table" "test" {
{{- template "region" }}
  vpc_id = aws_vpc.test.id
}
