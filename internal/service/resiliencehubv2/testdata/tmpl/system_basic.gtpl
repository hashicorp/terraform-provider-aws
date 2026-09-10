resource "aws_resiliencehubv2_system" "test" {
{{- template "region" }}
  name = var.rName

{{- template "tags" . }}
}
