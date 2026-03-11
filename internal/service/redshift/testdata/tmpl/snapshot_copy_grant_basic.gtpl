resource "aws_redshift_snapshot_copy_grant" "test" {
{{- template "region" }}
  snapshot_copy_grant_name = var.rName
{{- template "tags" . }}
}
