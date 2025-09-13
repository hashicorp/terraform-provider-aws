resource "aws_media_packagev2_channel_group" "test" {
  name = var.rName
{{- template "tags" . }}
}