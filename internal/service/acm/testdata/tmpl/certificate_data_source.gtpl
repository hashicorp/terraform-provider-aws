data "aws_acm_certificate" "test" {
  domain = aws_acm_certificate.test.domain_name
}
