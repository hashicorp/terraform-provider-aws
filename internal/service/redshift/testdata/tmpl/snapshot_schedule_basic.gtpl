resource "aws_redshift_snapshot_schedule" "test" {
{{- template "region" }}
  identifier = var.rName
  definitions = [
    "rate(12 hours)",
  ]
{{- template "tags" . }}
}
