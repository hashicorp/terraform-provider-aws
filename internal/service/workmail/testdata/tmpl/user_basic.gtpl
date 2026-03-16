resource "aws_workmail_organization" "test" {
{{- template "region" }}
  organization_alias = var.rName
  delete_directory   = true
}

resource "aws_workmail_user" "test" {
{{- template "region" }}
  organization_id = aws_workmail_organization.test.organization_id
  name            = var.rName
  display_name    = var.rName
  password        = "TestTest1234!"
}