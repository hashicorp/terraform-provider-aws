resource "aws_ssm_maintenance_window" "test" {
{{- template "region" }}
  name     = var.rName
  cutoff   = 1
  duration = 3
  schedule = "cron(0 16 ? * TUE *)"

{{- template "tags" . }}
}
