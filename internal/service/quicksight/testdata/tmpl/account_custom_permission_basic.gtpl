resource "aws_quicksight_custom_permissions" "test" {
{{ template "region" }}
  custom_permissions_name = var.rName

  capabilities {
    print_reports    = "DENY"
    share_dashboards = "DENY"
  }
}

resource "aws_quicksight_account_custom_permission" "test" {
{{ template "region" }}
  custom_permissions_name = aws_quicksight_custom_permissions.test.custom_permissions_name
}
