resource "aws_workmail_organization" "test" {
{{- template "region" }}
  organization_alias = var.rName
  delete_directory   = true
}

resource "aws_workmail_domain" "test" {
{{- template "region" }}
  organization_id = aws_workmail_organization.test.organization_id
  domain_name     = "${var.rName}.example.com"
}
