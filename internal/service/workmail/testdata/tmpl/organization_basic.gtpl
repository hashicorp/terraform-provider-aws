resource "aws_workmail_organization" "test" {
{{- template "region" }}
  organization_alias = var.rName
  delete_directory   = true
{{- template "tags" . }}
}
