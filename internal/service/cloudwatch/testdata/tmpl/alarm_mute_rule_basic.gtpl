resource "aws_cloudwatch_alarm_mute_rule" "test" {
  name = var.rName

  rule {
    schedule {
      duration   = "PT4H"
      expression = "cron(0 2 * * *)"
    }
  }

{{- template "tags" . }}
}
