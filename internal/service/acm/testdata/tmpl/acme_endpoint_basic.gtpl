resource "aws_acm_acme_endpoint" "test" {
{{- template "region" }}
  authorization_behavior = "PRE_APPROVED"

  certificate_authority {
    public_certificate_authority {
      allowed_key_algorithms = ["RSA_2048"]
    }
  }
}
