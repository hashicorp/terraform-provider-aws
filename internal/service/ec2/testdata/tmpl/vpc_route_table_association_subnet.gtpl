resource "aws_route_table_association" "test" {
{{- template "region" }}
  route_table_id = aws_route_table.test.id
  subnet_id      = aws_subnet.test.id
}

resource "aws_route_table" "test" {
{{- template "region" }}
  vpc_id = aws_vpc.test.id

  route {
    cidr_block = "10.0.0.0/8"
    gateway_id = aws_internet_gateway.test.id
  }
}

# testAccRouteTableAssociationConfigBaseVPC

resource "aws_vpc" "test" {
{{- template "region" }}
  cidr_block = "10.1.0.0/16"
}

resource "aws_subnet" "test" {
{{- template "region" }}
  vpc_id     = aws_vpc.test.id
  cidr_block = "10.1.1.0/24"
}

resource "aws_internet_gateway" "test" {
{{- template "region" }}
  vpc_id = aws_vpc.test.id
}
