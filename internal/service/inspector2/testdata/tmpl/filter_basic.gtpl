resource "aws_inspector2_filter" "test" {
{{- template "region" }}
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
