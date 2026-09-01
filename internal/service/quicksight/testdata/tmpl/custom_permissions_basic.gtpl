resource "aws_quicksight_custom_permissions" "test" {
  custom_permissions_name = var.rName

  capabilities {
    print_reports    = "DENY"
    share_dashboards = "DENY"
  }
{{- template "tags" . }}
}