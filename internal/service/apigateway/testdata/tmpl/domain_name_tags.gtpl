resource "aws_api_gateway_domain_name" "test" {
  domain_name              = var.rName
  regional_certificate_arn = aws_acm_certificate.test.arn

  endpoint_configuration {
    types = ["REGIONAL"]
  }

{{- template "tags" . }}
}

resource "aws_acm_certificate" "test" {
  certificate_body = var.certificate_pem
  private_key      = var.private_key_pem
}
