resource "aws_ssoadmin_trusted_token_issuer" "test" {
{{- template "region" }}
  name                      = var.rName
  instance_arn              = tolist(data.aws_ssoadmin_instances.test.arns)[0]
  trusted_token_issuer_type = "OIDC_JWT"

  trusted_token_issuer_configuration {
    oidc_jwt_configuration {
      claim_attribute_path          = "email"
      identity_store_attribute_path = "emails.value"
      issuer_url                    = "https://example.com"
      jwks_retrieval_option         = "OPEN_ID_DISCOVERY"
    }
  }
{{- template "tags" . }}
}

data "aws_ssoadmin_instances" "test" {
{{- template "region" -}}
}
