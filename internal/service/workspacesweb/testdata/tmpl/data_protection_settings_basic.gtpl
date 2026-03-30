resource "aws_workspacesweb_data_protection_settings" "test" {
  display_name = "test"

{{- template "tags" . }}

}