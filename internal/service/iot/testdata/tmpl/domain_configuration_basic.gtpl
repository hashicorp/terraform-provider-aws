resource "aws_acm_certificate" "test" {
{{- template "region" }}
  domain_name       = var.domainName
  validation_method = "DNS"
}

data "aws_route53_zone" "test" {
  name         = var.rootDomain
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
{{- template "region" }}
  depends_on      = [aws_route53_record.test]
  certificate_arn = aws_acm_certificate.test.arn
}

resource "aws_iot_domain_configuration" "test" {
{{- template "region" }}
  depends_on              = [aws_acm_certificate_validation.test]
  name                    = var.rName
  domain_name             = var.domainName
  server_certificate_arns = [aws_acm_certificate.test.arn]
{{- template "tags" . }}
}
