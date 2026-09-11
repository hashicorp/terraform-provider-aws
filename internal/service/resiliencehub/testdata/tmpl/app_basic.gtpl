resource "aws_resiliencehub_app" "test" {
{{- template "region" }}
  name = var.rName
{{- template "tags" . }}
}
