resource "aws_mediapackagevod_packaging_group" "test" {
  name = var.rName
{{- template "tags" . }}
}