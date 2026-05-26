resource "aws_resiliencehubv2_system" "test" {
  name = var.rName

{{- template "tags" . }}
}
