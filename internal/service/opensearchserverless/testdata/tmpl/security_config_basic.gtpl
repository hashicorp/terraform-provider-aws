resource "aws_opensearchserverless_security_config" "test" {
{{- template "region" }}
  name = var.rName
  type = "saml"
  saml_options {
    metadata = file("test-fixtures/idp-metadata.xml")
  }

{{- template "tags" . }}
}
