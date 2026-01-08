{{ template "acctest.ConfigVPCWithSubnets" 1 }}

resource "aws_route_table_association" "test" {
{{- template "region" }}
  route_table_id = aws_route_table.test.id
  subnet_id      = aws_subnet.test[0].id
{{- template "tags" }}
}

resource "aws_internet_gateway" "test" {
{{- template "region" }}
  vpc_id = aws_vpc.test.id
}

resource "aws_route_table" "test" {
{{- template "region" }}
  vpc_id = aws_vpc.test.id

  route {
    cidr_block = "10.0.0.0/8"
    gateway_id = aws_internet_gateway.test.id
  }
}
