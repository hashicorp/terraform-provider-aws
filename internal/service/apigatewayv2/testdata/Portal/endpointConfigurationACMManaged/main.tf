# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_apigatewayv2_portal" "test" {
  authorization {
    none {}
  }

  endpoint_configuration {
    acm_managed {
      certificate_arn = aws_acm_certificate_validation.test.certificate_arn
      domain_name     = aws_acm_certificate.test.domain_name
    }
  }

  portal_content {
    display_name = var.rName

    theme {
      custom_colors {
        accent_color           = "#FF9900"
        background_color       = "#FFFFFF"
        error_validation_color = "#D13212"
        header_color           = "#232F3E"
        navigation_color       = "#232F3E"
        text_color             = "#16191F"
      }
    }
  }
}

resource "aws_acm_certificate" "test" {
  domain_name       = var.domain_name
  validation_method = "DNS"
}

data "aws_route53_zone" "test" {
  name         = var.root_domain
  private_zone = false
}

resource "aws_route53_record" "test" {
  allow_overwrite = true
  name            = tolist(aws_acm_certificate.test.domain_validation_options)[0].resource_record_name
  records         = [tolist(aws_acm_certificate.test.domain_validation_options)[0].resource_record_value]
  ttl             = 60
  type            = tolist(aws_acm_certificate.test.domain_validation_options)[0].resource_record_type
  zone_id         = data.aws_route53_zone.test.zone_id
}

resource "aws_acm_certificate_validation" "test" {
  certificate_arn         = aws_acm_certificate.test.arn
  validation_record_fqdns = [aws_route53_record.test.fqdn]
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "root_domain" {
  description = "Route 53 zone of the custom domain"
  type        = string
  nullable    = false
}

variable "domain_name" {
  description = "Custom domain name"
  type        = string
  nullable    = false
}
