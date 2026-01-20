resource "aws_vpc_endpoint" "test" {
{{- template "region" }}
  vpc_id       = aws_vpc.test.id
  service_name = "com.amazonaws.${data.aws_region.current.region}.s3"
{{- template "tags" }}
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
