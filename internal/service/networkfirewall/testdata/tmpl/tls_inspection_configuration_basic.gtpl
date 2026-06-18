resource "aws_networkfirewall_tls_inspection_configuration" "test" {
{{- template "region" }}
  name = var.rName

  tls_inspection_configuration {
    server_certificate_configuration {
      server_certificate {
        resource_arn = aws_acm_certificate.test.arn
      }
      scope {
        protocols = [6]
        destination {
          address_definition = "0.0.0.0/0"
        }
      }
    }
  }
{{- template "tags" . }}
}

# testAccTLSInspectionConfigurationConfig_certificateBase

resource "aws_acmpca_certificate_authority" "test" {
{{- template "region" }}
  permanent_deletion_time_in_days = 7
  type                            = "ROOT"

  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = var.common_name
    }
  }
}

resource "aws_acmpca_certificate" "test" {
{{- template "region" }}
  certificate_authority_arn   = aws_acmpca_certificate_authority.test.arn
  certificate_signing_request = aws_acmpca_certificate_authority.test.certificate_signing_request
  signing_algorithm           = "SHA512WITHRSA"

  template_arn = "arn:${data.aws_partition.current.partition}:acm-pca:::template/RootCACertificate/V1"

  validity {
    type  = "YEARS"
    value = 2
  }
}

resource "aws_acmpca_certificate_authority_certificate" "test" {
{{- template "region" }}
  certificate_authority_arn = aws_acmpca_certificate_authority.test.arn

  certificate       = aws_acmpca_certificate.test.certificate
  certificate_chain = aws_acmpca_certificate.test.certificate_chain
}

data "aws_partition" "current" {}

resource "aws_acm_certificate" "test" {
{{- template "region" }}
  domain_name               = var.certificate_domain
  certificate_authority_arn = aws_acmpca_certificate_authority.test.arn

  depends_on = [
    aws_acmpca_certificate_authority_certificate.test,
  ]
}
