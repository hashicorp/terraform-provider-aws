# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_acm_certificate_validation" "test" {
  depends_on = [aws_route53_record.test]

  certificate_arn = aws_acm_certificate.test.arn
}

resource "aws_acm_certificate" "test" {
  domain_name       = var.domainName
  validation_method = "DNS"
}

data "aws_route53_zone" "test" {
  name         = var.rootDomain
  private_zone = false
}

resource "aws_route53_record" "test" {
  name    = tolist(aws_acm_certificate.test.domain_validation_options)[0].resource_record_name
  records = [tolist(aws_acm_certificate.test.domain_validation_options)[0].resource_record_value]
  ttl     = 60
  type    = tolist(aws_acm_certificate.test.domain_validation_options)[0].resource_record_type
  zone_id = data.aws_route53_zone.test.zone_id
}


variable "domainName" {
  type     = string
  nullable = false
}

variable "rootDomain" {
  type     = string
  nullable = false
}
