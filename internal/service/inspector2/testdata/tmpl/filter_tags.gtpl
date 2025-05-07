resource "aws_inspector2_filter" "test" {
  name   = var.rName
  action = "NONE"
  filter_criteria {
    aws_account_id {
      comparison = "EQUALS"
      value      = "111222333444"
    }
  }

{{- template "tags" . }}
}
