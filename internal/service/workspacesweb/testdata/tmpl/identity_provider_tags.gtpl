resource "aws_workspacesweb_portal" "test" {
  display_name = "test"
}

resource "aws_workspacesweb_identity_provider" "test" {
  identity_provider_name = "test"
  identity_provider_type = "SAML"
  portal_arn            = aws_workspacesweb_portal.test.portal_arn

  identity_provider_details = {
    MetadataURL = "https://example.com/metadata"
  }

{{- template "tags" . }}

}
