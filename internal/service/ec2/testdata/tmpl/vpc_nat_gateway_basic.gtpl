resource "aws_nat_gateway" "test" {
{{- template "region" }}
  connectivity_type = "private"
  subnet_id         = aws_subnet.private.id
}

resource "aws_vpc" "test" {
{{- template "region" }}
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = var.rName
  }
}

resource "aws_subnet" "private" {
{{- template "region" }}
  vpc_id                  = aws_vpc.test.id
  cidr_block              = "10.0.1.0/24"
  map_public_ip_on_launch = false

  tags = {
    Name = var.rName
  }
}
