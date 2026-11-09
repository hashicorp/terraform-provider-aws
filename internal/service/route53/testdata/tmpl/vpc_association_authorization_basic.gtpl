resource "aws_route53_vpc_association_authorization" "test" {
{{- template "region" }}
  zone_id = aws_route53_zone.test.id
  vpc_id  = aws_vpc.alternate.id
}

resource "aws_vpc" "alternate" {
  provider             = "awsalternate"
  cidr_block           = cidrsubnet("10.0.0.0/8", 8, 0)
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = var.rName
  }
}

resource "aws_route53_zone" "test" {
  name = "${var.rName}.example.com"

  vpc {
    vpc_id = aws_vpc.test.id
  }
}

resource "aws_vpc" "test" {
  cidr_block           = cidrsubnet("10.0.0.0/8", 8, 1)
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = var.rName
  }
}

{{ template "acctest.ConfigAlternateAccountProvider" }}
