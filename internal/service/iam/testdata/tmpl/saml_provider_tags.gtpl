resource "aws_iam_saml_provider" "test" {
{{- template "region" }}
  name                   = var.rName
  saml_metadata_document = templatefile("./test-fixtures/saml-metadata.xml.tpl", { entity_id = "https://example.com" })
{{- template "tags" }}
}
