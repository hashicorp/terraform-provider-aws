resource "aws_kms_external_key" "test" {
  description             = var.rName
  deletion_window_in_days = 7

{{- template "tags" . }}
}
