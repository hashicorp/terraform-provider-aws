resource "aws_route53_resolver_rule_association" "test" {
{{- template "region" }}
  name             = var.rName
  resolver_rule_id = aws_route53_resolver_rule.test.id
  vpc_id           = aws_vpc.test.id
}

resource "aws_vpc" "test" {
{{- template "region" }}
  cidr_block           = "10.6.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true
}

resource "aws_route53_resolver_rule" "test" {
{{- template "region" }}
  domain_name = var.domain
  name        = var.rName
  rule_type   = "SYSTEM"
}
