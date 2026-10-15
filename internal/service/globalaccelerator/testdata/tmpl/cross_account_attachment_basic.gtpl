resource "aws_globalaccelerator_cross_account_attachment" "test" {
{{- template "region" }}
  name = var.rName
{{- template "tags" . }}
}
