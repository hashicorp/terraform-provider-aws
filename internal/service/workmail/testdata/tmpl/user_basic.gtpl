resource "aws_workmail_organization" "test" {
{{- template "region" }}
  organization_alias = var.rName
  delete_directory   = true
}

resource "aws_workmail_user" "test" {
{{- template "region" }}
  organization_id = aws_workmail_organization.test.organization_id
  email           = "${var.rName}@${aws_workmail_organization.test.default_mail_domain}"
  name            = var.rName
  display_name    = var.rName
  city            = "bangalore"
  office          = "hashicorp"
}