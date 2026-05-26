resource "aws_vpc_endpoint_route_table_association" "test" {
{{- template "region" }}
  vpc_endpoint_id = aws_vpc_endpoint.test.id
  route_table_id  = aws_route_table.test.id
}

resource "aws_vpc" "test" {
{{- template "region" }}
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = var.rName
  }
}

data "aws_region" "current" {
{{- template "region" }}
}

resource "aws_vpc_endpoint" "test" {
{{- template "region" }}
  vpc_id       = aws_vpc.test.id
  service_name = "com.amazonaws.${data.aws_region.current.region}.s3"

  tags = {
    Name = var.rName
  }
}

resource "aws_route_table" "test" {
{{- template "region" }}
  vpc_id = aws_vpc.test.id

  tags = {
    Name = var.rName
  }
}
