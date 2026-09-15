resource "aws_networkmonitor_monitor" "test" {
  monitor_name = var.rName
{{- template "tags" . }}
}
