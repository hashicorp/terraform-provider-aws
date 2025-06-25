resource "aws_workspacesweb_ip_access_settings" "test" {
  display_name = "test"
  ip_rule {
    ip_range = "10.0.0.0/16"
  }

{{- template "tags" . }}

}