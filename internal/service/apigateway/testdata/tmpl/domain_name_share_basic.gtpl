resource "aws_api_gateway_domain_name_share" "test" {
{{- template "region" }}
  domain_name_id   = aws_api_gateway_domain_name.test.domain_name_id
  allowed_accounts = [data.aws_caller_identity.current.account_id]
}

data "aws_caller_identity" "current" {
}

resource "aws_api_gateway_domain_name" "test" {
{{- template "region" }}
  domain_name     = var.rName
  certificate_arn = aws_acm_certificate.test.arn

  endpoint_configuration {
    types = ["PRIVATE"]
  }
}

resource "aws_acm_certificate" "test" {
{{- template "region" }}
  certificate_body = var.certificate_pem
  private_key      = var.private_key_pem
}
