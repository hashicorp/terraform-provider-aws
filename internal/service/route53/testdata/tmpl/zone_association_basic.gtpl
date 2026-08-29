resource "aws_route53_zone_association" "test" {
{{- template "region" }}
  zone_id = aws_route53_zone.foo.id
  vpc_id  = aws_vpc.bar.id
}

resource "aws_vpc" "foo" {
  cidr_block           = "10.6.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = var.rName
  }
}

resource "aws_vpc" "bar" {
  cidr_block           = "10.7.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = var.rName
  }
}

resource "aws_route53_zone" "foo" {
  name = "${var.rName}.example.com"

  vpc {
    vpc_id = aws_vpc.foo.id
  }

  lifecycle {
    ignore_changes = [vpc]
  }
}
