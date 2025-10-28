resource "aws_ssm_maintenance_window_target" "test" {
{{- template "region" }}
  name          = var.rName
  description   = "This resource is for test purpose only"
  window_id     = aws_ssm_maintenance_window.test.id
  resource_type = "INSTANCE"

  targets {
    key    = "tag:Name"
    values = ["acceptance_test"]
  }

  targets {
    key    = "tag:Name2"
    values = ["acceptance_test", "acceptance_test2"]
  }
}

resource "aws_ssm_maintenance_window" "test" {
{{- template "region" }}
  name     = var.rName
  schedule = "cron(0 16 ? * TUE *)"
  duration = 3
  cutoff   = 1
}
