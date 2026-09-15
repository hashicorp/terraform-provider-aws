resource "aws_workmail_organization" "test" {
{{- template "region" }}
  organization_alias = var.rName
  delete_directory   = true
}

resource "aws_workmail_default_domain" "test" {
{{- template "region" }}
  organization_id = aws_workmail_organization.test.organization_id
  domain_name     = aws_workmail_organization.test.default_mail_domain
}