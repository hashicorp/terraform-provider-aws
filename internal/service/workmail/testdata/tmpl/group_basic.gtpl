resource "aws_workmail_group" "test" {
{{- template "region" }}
  organization_id = aws_workmail_organization.test.organization_id
  email           = "${var.rName}@${aws_workmail_organization.test.default_mail_domain}"
  name            = var.rName
}

resource "aws_workmail_organization" "test" {
{{- template "region" }}
  organization_alias = var.rName
  delete_directory   = true
}
