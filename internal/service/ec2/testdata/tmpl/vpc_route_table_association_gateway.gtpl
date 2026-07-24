resource "aws_route_table_association" "test" {
{{- template "region" }}
  route_table_id = aws_route_table.test.id
  gateway_id     = aws_vpn_gateway.test.id
}

resource "aws_route_table" "test" {
{{- template "region" }}
  vpc_id = aws_vpc.test.id

  route {
    cidr_block           = aws_subnet.test.cidr_block
    network_interface_id = aws_network_interface.test.id
  }
}

resource "aws_network_interface" "test" {
{{- template "region" }}
  subnet_id = aws_subnet.test.id
}

resource "aws_vpn_gateway" "test" {
{{- template "region" }}
  vpc_id = aws_vpc.test.id
}

resource "aws_vpc" "test" {
{{- template "region" }}
  cidr_block = "10.1.0.0/16"
}

resource "aws_subnet" "test" {
{{- template "region" }}
  vpc_id     = aws_vpc.test.id
  cidr_block = cidrsubnet(aws_vpc.test.cidr_block, 8, 1)
}
