resource "aws_workspacesweb_trust_store" "test" {
{{- template "region" }}
  certificate {
    body = aws_acmpca_certificate.test.certificate
  }
{{- template "tags" . }}
}

resource "aws_acmpca_certificate_authority" "test" {
{{- template "region" }}
  permanent_deletion_time_in_days = 7

  type = "ROOT"

  certificate_authority_configuration {
    key_algorithm     = "RSA_2048"
    signing_algorithm = "SHA256WITHRSA"

    subject {
      common_name = "example.com"
    }
  }
}

resource "aws_acmpca_certificate" "test" {
{{- template "region" }}
  certificate_authority_arn   = aws_acmpca_certificate_authority.test.arn
  certificate_signing_request = aws_acmpca_certificate_authority.test.certificate_signing_request
  signing_algorithm           = "SHA256WITHRSA"

  template_arn = "arn:aws:acm-pca:::template/RootCACertificate/V1"

  validity {
    type  = "YEARS"
    value = 1
  }
}
