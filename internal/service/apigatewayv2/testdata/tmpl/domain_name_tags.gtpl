resource "aws_apigatewayv2_domain_name" "test" {
  domain_name = var.rName

  domain_name_configuration {
    certificate_arn = aws_acm_certificate.test.arn
    endpoint_type   = "REGIONAL"
    security_policy = "TLS_1_2"
  }

{{- template "tags" . }}
}

resource "aws_acm_certificate" "test" {
  certificate_body = var.certificate_pem
  private_key      = var.private_key_pem
}
