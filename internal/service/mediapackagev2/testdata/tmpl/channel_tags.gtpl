resource "aws_media_packagev2_channel_group" "test" {
  name = var.rName
{{- template "tags" . }}
}

resource "aws_media_packagev2_channel" "test" {
  channel_group_name = aws_media_packagev2_channel_group.test.name
  name               = var.rName
{{- template "tags" . }}
}
